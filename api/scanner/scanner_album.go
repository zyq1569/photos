package scanner

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/scanner/scanner_tasks"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func NewRootAlbum(db *gorm.DB, rootPath string, owner *models.User) (*models.Album, error) {

	if !ValidRootPath(rootPath) {
		// if runtime.GOOS == "windows" {
		path, err := GetCurrentPath()
		if err == nil {
			return nil, ErrorInvalidRootPath
		} else {
			rootPath = path + "photos"
		}
		log.Printf("Warn: set rootPath : '%s'\n%s\n", rootPath, "default")
		// }
	}
	if runtime.GOOS != "windows" {
		if !path.IsAbs(rootPath) {
			wd, err := os.Getwd()
			if err != nil {
				return nil, err
			}
			rootPath = path.Join(wd, rootPath)
		}
	}

	owners := []models.User{
		*owner,
	}

	var matchedAlbums []models.Album
	if err := db.Where("path_hash = ?", models.MD5Hash(rootPath)).Find(&matchedAlbums).Error; err != nil {
		return nil, err
	}

	if len(matchedAlbums) > 0 {
		album := matchedAlbums[0]

		var matchedUserAlbumCount int64
		if err := db.Table("user_albums").Where("user_id = ?", owner.ID).Where("album_id = ?", album.ID).Count(&matchedUserAlbumCount).Error; err != nil {
			return nil, err
		}

		if matchedUserAlbumCount > 0 {
			return nil, errors.New(fmt.Sprintf("user already owns a path containing this path: %s", rootPath))
		}

		if err := db.Model(&owner).Association("Albums").Append(&album); err != nil {
			return nil, errors.Wrap(err, "add owner to already existing album")
		}

		return &album, nil
	} else {
		album := models.Album{
			Title:  path.Base(rootPath),
			Path:   rootPath,
			Owners: owners,
		}

		if err := db.Create(&album).Error; err != nil {
			return nil, err
		}

		return &album, nil
	}
}

var ErrorInvalidRootPath = errors.New("invalid root path")

func ValidRootPath(rootPath string) bool {
	_, err := os.Stat(rootPath)
	if err != nil {
		log.Printf("Warn: invalid root path: '%s'\n%s\n", rootPath, err)
		return false
	}

	return true
}

func ScanAlbum(ctx scanner_task.TaskContext) error {

	newCtx, err := scanner_tasks.Tasks.BeforeScanAlbum(ctx)
	if err != nil {
		return errors.Wrapf(err, "before scan album (%s)", ctx.GetAlbum().Path)
	}
	ctx = newCtx

	// Scan for photos
	albumMedia, err := findMediaForAlbum(ctx)
	if err != nil {
		return errors.Wrapf(err, "find media for album (%s): %s", ctx.GetAlbum().Path, err)
	}

	changedMedia := make([]*models.Media, 0)
	for i, media := range albumMedia {
		updatedURLs := []*models.MediaURL{}

		mediaData := media_encoding.NewEncodeMediaData(media)

		// define new ctx for scope of for-loop
		ctx, err := scanner_tasks.Tasks.BeforeProcessMedia(ctx, &mediaData)
		if err != nil {
			return err
		}

		transactionError := ctx.DatabaseTransaction(func(ctx scanner_task.TaskContext) error {
			updatedURLs, err = processMedia(ctx, &mediaData)
			if err != nil {
				return errors.Wrapf(err, "process media (%s)", media.Path)
			}

			if len(updatedURLs) > 0 {
				changedMedia = append(changedMedia, media)
			}

			return nil
		})

		if transactionError != nil {
			return errors.Wrap(err, "process media database transaction")
		}

		if err = scanner_tasks.Tasks.AfterProcessMedia(ctx, &mediaData, updatedURLs, i, len(albumMedia)); err != nil {
			return errors.Wrap(err, "after process media")
		}
	}

	if err := scanner_tasks.Tasks.AfterScanAlbum(ctx, changedMedia, albumMedia); err != nil {
		return errors.Wrap(err, "after scan album")
	}

	return nil
}

func findMediaForAlbum(ctx scanner_task.TaskContext) ([]*models.Media, error) {

	albumMedia := make([]*models.Media, 0)

	dirContent, err := ioutil.ReadDir(ctx.GetAlbum().Path)
	if err != nil {
		return nil, err
	}

	for _, item := range dirContent {
		mediaPath := path.Join(ctx.GetAlbum().Path, item.Name())

		isDirSymlink, err := utils.IsDirSymlink(mediaPath)
		if err != nil {
			log.Printf("Cannot detect whether %s is symlink to a directory. Pretending it is not", mediaPath)
			isDirSymlink = false
		}

		if !item.IsDir() && !isDirSymlink && ctx.GetCache().IsPathMedia(mediaPath) {
			skip, err := scanner_tasks.Tasks.MediaFound(ctx, item, mediaPath)
			if err != nil {
				return nil, err
			}
			if skip {
				continue
			}

			err = ctx.DatabaseTransaction(func(ctx scanner_task.TaskContext) error {
				media, isNewMedia, err := ScanMedia(ctx.GetDB(), mediaPath, ctx.GetAlbum().ID, ctx.GetCache())
				if err != nil {
					return errors.Wrapf(err, "scanning media error (%s)", mediaPath)
				}

				if err = scanner_tasks.Tasks.AfterMediaFound(ctx, media, isNewMedia); err != nil {
					return err
				}

				albumMedia = append(albumMedia, media)

				return nil
			})

			if err != nil {
				scanner_utils.ScannerError("Error scanning media for album (%d): %s\n", ctx.GetAlbum().ID, err)
				continue
			}
		}

	}

	return albumMedia, nil
}

func processMedia(ctx scanner_task.TaskContext, mediaData *media_encoding.EncodeMediaData) ([]*models.MediaURL, error) {

	// Make sure media cache directory exists
	mediaCachePath, err := mediaData.Media.CachePath()
	if err != nil {
		return []*models.MediaURL{}, errors.Wrap(err, "cache directory error")
	}

	return scanner_tasks.Tasks.ProcessMedia(ctx, mediaData, mediaCachePath)
}

func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	//fmt.Println("path111:", path)
	if runtime.GOOS == "windows" {
		path = strings.Replace(path, "\\", "/", -1)
	}
	//fmt.Println("path222:", path)
	i := strings.LastIndex(path, "/")
	if i < 0 {
		return "", errors.New(`Can't find "/" or "\".`)
	}
	//fmt.Println("path333:", path)
	return string(path[0 : i+1]), nil
}
