#include "mainwindow.h"
#include "ui_mainwindow.h"

#include<QDir>

#ifndef QT_NO_SYSTEMTRAYICON
#include <windows.h>
#endif

#include <QtWidgets>

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , ui(new Ui::MainWindow)
{
    ui->setupUi(this);

    createTrayIcon();
    connect(m_TrayIcon, &QSystemTrayIcon::activated, this, &MainWindow::iconActivated);
    m_pQProcess = new QProcess(this);
}

MainWindow::~MainWindow()
{
    QString Dir = QDir::currentPath();
    m_pQProcess->close();
    delete  m_pQProcess;
    m_pQProcess = NULL;
    delete ui;
}


void MainWindow::on_startButton_clicked()
{
    QString program = QDir::currentPath() +"/Photoview.exe" ;
    QString text = ui->startButton->text();
    if (text == "Start")
    {
        QString program = QDir::currentPath() +"/Photoview.exe" ;
        QStringList arg ;
        m_pQProcess->start(program,arg);
        ui->startButton->setText("Stop");
    }
    else
    {
        ui->startButton->setText("Start");
        m_pQProcess->close();
    }
}

void MainWindow::createTrayIcon()
{
    m_RestoreAction = new QAction(tr("&Restore"), this);
    connect(m_RestoreAction, &QAction::triggered, this, &QWidget::showNormal);

    m_QuitAction = new QAction(tr("&Quit"), this);
    connect(m_QuitAction, &QAction::triggered, qApp, &QCoreApplication::quit);

    m_TrayIconMenu = new QMenu(this);
    m_TrayIconMenu->addAction(m_RestoreAction);
    m_TrayIconMenu->addSeparator();
    m_TrayIconMenu->addAction(m_QuitAction);

    m_Icon = QIcon(":/images/server.png"), tr("Server");
    m_TrayIcon = new QSystemTrayIcon(this);
    m_TrayIcon->setIcon(m_Icon);
    setWindowIcon(m_Icon);
    m_TrayIcon->setToolTip("Healthsystem ServerManage UI");
    m_TrayIcon->setContextMenu(m_TrayIconMenu);
    m_TrayIcon->show();
}


void MainWindow::closeEvent(QCloseEvent *event)
{
    QCoreApplication::quit();
}
void MainWindow::changeEvent(QEvent *event)
{
    if(event->type()!=QEvent::WindowStateChange)
        return;
    if(this->windowState()==Qt::WindowMinimized)
    {
#ifndef QT_NO_SYSTEMTRAYICON
        if (m_TrayIcon->isVisible())
        {
            //            QMessageBox::information(this, tr("Systray"),
            //                                     tr("The program will keep running in the "
            //                                        "system tray. To terminate the program, "
            //                                        "choose <b>close</b> in the context menu "
            //                                        "of the system tray entry."));
            hide();
            event->ignore();
        }
#endif
    }
}

#ifndef QT_NO_SYSTEMTRAYICON
void MainWindow::iconActivated(QSystemTrayIcon::ActivationReason reason)
{
    switch (reason)
    {

    case QSystemTrayIcon::Trigger:
    case QSystemTrayIcon::DoubleClick:
        /// 1. set Window TOPMOST 2. set Window NOTOPMOST
        this->showNormal();
        ::SetWindowPos(HWND(this->winId()), HWND_TOPMOST, 0, 0, 0, 0, SWP_NOMOVE | SWP_NOSIZE | SWP_SHOWWINDOW);
        ::SetWindowPos(HWND(this->winId()), HWND_NOTOPMOST, 0, 0, 0, 0, SWP_NOMOVE | SWP_NOSIZE | SWP_SHOWWINDOW);
        this->show();
        break;
    case QSystemTrayIcon::MiddleClick:
    {
        QSystemTrayIcon::MessageIcon msgIcon = QSystemTrayIcon::MessageIcon(1/*MessageIcon::Information*/);
        m_TrayIcon->showMessage("HServerManageUI", "This is a HServerManageUI(PACS&RIS) App!", msgIcon,500);
        break;
    }

    default:
        ;
    }
}
#endif
