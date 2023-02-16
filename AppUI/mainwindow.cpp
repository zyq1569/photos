#include "mainwindow.h"
#include "ui_mainwindow.h"

#include<QDir>


MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , ui(new Ui::MainWindow)
{
    ui->setupUi(this);
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
