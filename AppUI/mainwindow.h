#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QMainWindow>

QT_BEGIN_NAMESPACE
namespace Ui { class MainWindow; }
QT_END_NAMESPACE

#include<QProcess>

///----only windows
#ifndef QT_NO_SYSTEMTRAYICON
#include <QSystemTrayIcon>
#else
class QSystemTrayIcon;
#endif
///
///
#include <QMainWindow>
#include<QStandardItemModel>
class QSystemTrayIcon;
class MainWindow : public QMainWindow
{
    Q_OBJECT

public:
    MainWindow(QWidget *parent = nullptr);
    ~MainWindow();
private:
    QProcess *m_pQProcess;
private:
    QSystemTrayIcon *m_TrayIcon;
    QMenu *m_TrayIconMenu;
    QAction *m_RestoreAction;
    QAction *m_QuitAction;
    QIcon m_Icon;
protected:
    void changeEvent(QEvent * event);
    void createTrayIcon();
    void closeEvent(QCloseEvent *event);


#ifndef QT_NO_SYSTEMTRAYICON
    void iconActivated(QSystemTrayIcon::ActivationReason reason);
#endif

private slots:
    void on_startButton_clicked();

private:
    Ui::MainWindow *ui;
};
#endif // MAINWINDOW_H
