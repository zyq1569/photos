package main

import (
	"log"
	// "mime"
	"net/http"
	"path"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/joho/godotenv"

	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/dataloader"
	"github.com/photoview/photoview/api/graphql/auth"
	graphql_endpoint "github.com/photoview/photoview/api/graphql/endpoint"
	"github.com/photoview/photoview/api/routes"
	"github.com/photoview/photoview/api/scanner/exif"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
	"github.com/photoview/photoview/api/scanner/periodic_scanner"
	"github.com/photoview/photoview/api/scanner/scanner_queue"
	"github.com/photoview/photoview/api/server"
	"github.com/photoview/photoview/api/utils"

	"github.com/99designs/gqlgen/graphql/playground"
	// log4go "github.com/jeanphorn/log4go"
)

func main() {

	// log4go.AddFilter("stdout", log4go.ERROR, log4go.NewConsoleLogWriter())
	// log4go.AddFilter("file", log4go.ERROR, log4go.NewFileLogWriter("./log/PhotosApp.log", false, true))
	// log4go.Info("----------Level:ERROR---------------")
	// Windows may be missing this
	// mime.AddExtensionType(".js", "application/javascript")

	// And then you create the FileServer like you normally would
	// http.Handle("/", http.FileServer(http.Dir("static")))
	log.Println("------------Starting Photoview...------------------")

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	devMode := utils.DevelopmentMode()

	db, err := database.SetupDatabase()
	if err != nil {
		log.Panicf("Could not connect to database: %s\n", err)
	}

	// Migrate database
	if err := database.MigrateDatabase(db); err != nil {
		log.Panicf("Could not migrate database: %s\n", err)
	}

	if err := scanner_queue.InitializeScannerQueue(db); err != nil {
		log.Panicf("Could not initialize scanner queue: %s\n", err)
	}

	if err := periodic_scanner.InitializePeriodicScanner(db); err != nil {
		log.Panicf("Could not initialize periodic scanner: %s", err)
	}

	executable_worker.InitializeExecutableWorkers()

	exif.InitializeEXIFParser()

	if err := face_detection.InitializeFaceDetector(db); err != nil {
		log.Panicf("Could not initialize face detector: %s\n", err)
	}

	rootRouter := mux.NewRouter()

	rootRouter.Use(dataloader.Middleware(db))
	rootRouter.Use(auth.Middleware(db))
	rootRouter.Use(server.LoggingMiddleware)
	rootRouter.Use(server.CORSMiddleware(devMode))

	apiListenURL := utils.ApiListenUrl()

	endpointRouter := rootRouter.PathPrefix(apiListenURL.Path).Subrouter()

	if devMode {
		endpointRouter.Handle("/", playground.Handler("GraphQL playground", path.Join(apiListenURL.Path, "/graphql")))
	} else {
		endpointRouter.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("photoview api endpoint"))
		})
	}

	endpointRouter.Handle("/graphql", graphql_endpoint.GraphqlEndpoint(db))

	photoRouter := endpointRouter.PathPrefix("/photo").Subrouter()
	routes.RegisterPhotoRoutes(db, photoRouter)

	videoRouter := endpointRouter.PathPrefix("/video").Subrouter()
	routes.RegisterVideoRoutes(db, videoRouter)

	downloadsRouter := endpointRouter.PathPrefix("/download").Subrouter()
	routes.RegisterDownloadRoutes(db, downloadsRouter)

	shouldServeUI := utils.ShouldServeUI()

	if shouldServeUI {
		spa := routes.NewSpaHandler(utils.UIPath(), "index.html")
		rootRouter.PathPrefix("/").Handler(spa)
	}

	if devMode {
		log.Printf("🚀 Graphql playground ready at %s\n", apiListenURL.String())
	} else {
		log.Printf("Photoview API endpoint listening at %s\n", apiListenURL.String())

		apiEndpoint := utils.ApiEndpointUrl()
		log.Printf("Photoview API public endpoint ready at %s\n", apiEndpoint.String())

		if uiEndpoint := utils.UiEndpointUrl(); uiEndpoint != nil {
			log.Printf("Photoview UI public endpoint ready at %s\n", uiEndpoint.String())
		} else {
			log.Println("Photoview UI public endpoint ready at /")
		}

		if !shouldServeUI {
			log.Printf("Notice: UI is not served by the the api (%s=0)", utils.EnvServeUI.GetName())
		}

	}

	log.Panic(http.ListenAndServe(":"+apiListenURL.Port(), handlers.CompressHandler(rootRouter)))
}

//https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Basics_of_HTTP/MIME_types
///freeget.one@gmail.com
//https://github.com/photoview/photoview
//
//sudo docker run -it --entrypoint sh viktorstrate/photoview
//copy
//sudo docker cp 74ba34a8do19ec:/app  /home/zyq/Downloads
//sudo docker cp /home/zyq/Downloads/dist 74ba34a8do19ec:/ui
//sudo docker cp 74ba34a819ec:/ui /home/zyq/Downloads/dist
//mount???
//sudo docker run -d -it --name mnt -v /home/Downloads/mnt:/ui viktorstrate/photoview:edge
//GRANT ALL PRIVILEGES ON *.* TO 'dbuser'@'%' IDENTIFIED BY 'dbuser' WITH GRANT OPTION;
//FLUSH PRIVILEGES;
//grant all privileges on `数据库名`.* to '用户名'@'%' identified by '密码' with grant option;
//grant all privileges on `photoview`.* to 'dbuser'@'%' identified by 'dbuser' with grant option;
//********
// Build the Web user-interface
// $ cd ui/
// $ npm install
// $ npm run build
// This builds the UI source code and saves it in the ui/build/ directory.

//docker run -d -P --name mariadb_connect -e MYSQL_ROOT_PASSWORD=dockersql  mariadb:10.5

///curl -fsSL https://deb.nodesource.com/setup_16.x | sudo -E bash - &&\ apt-get install -y nodejs
//https://codeandlife.com/2022/02/12/fixing-mime-type-golang-net-http-fileserver/
//https://vimsky.com/examples/detail/golang-ex-mime---AddExtensionType-function.html
