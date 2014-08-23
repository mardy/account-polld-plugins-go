/*
 Copyright 2014 Canonical Ltd.
 Authors: Sergio Schvezov <sergio.schvezov@canonical.com>

 This program is free software: you can redistribute it and/or modify it
 under the terms of the GNU General Public License version 3, as published
 by the Free Software Foundation.

 This program is distributed in the hope that it will be useful, but
 WITHOUT ANY WARRANTY; without even the implied warranties of
 MERCHANTABILITY, SATISFACTORY QUALITY, or FITNESS FOR A PARTICULAR
 PURPOSE.  See the GNU General Public License for more details.

 You should have received a copy of the GNU General Public License along
 with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

#include <QContactManager>
#include <QContactFilter>
#include <QContactEmailAddress>
#include <QContactDetailFilter>
#include <QContactManager>
#include <QContactAvatar>
#include <QCoreApplication>
#include <QScopedPointer>
#include <QDebug>
#include <QTimer>
#include <thread>

#include "qtcontacts.h"
#include "qtcontacts.hpp"
#include "qtcontacts.moc"

#ifdef __cplusplus
extern "C" {
#include "_cgo_export.h"
}
#endif

QTCONTACTS_USE_NAMESPACE

int mainloopStart() {
    static char empty[1] = {0};
    static char *argv[] = {empty, empty, empty};
    static int argc = 1;

    QCoreApplication mApp(argc, argv);
    qDebug() << "Starting mainloop";
    return mApp.exec();
}

void getAvatar(char *email) {
    QScopedPointer<Avatar> avatar(new Avatar());
    qDebug() << "Calling Avatar::getThumbnail";
    avatar->retrieveThumbnail(QString(email));
}

void Avatar::getThumbnail(char *email) {
    // set the map for this object and the passed email
    _signalMapper->setMapping(this, QString(email));

    QTimer::singleShot(0, this, SLOT(emitSignals()));
    QCoreApplication::instance()->processEvents();
    qDebug() << "Called processEvents";
}


void Avatar::emitSignals() {
    qDebug("Got emitSignals");
    emit readyToRetrieve();
}

void Avatar::retrieveThumbnail(const QString& email) {
    qDebug() << "Entering Avatar::retrieveThumbnail";
    QString avatar;

    QContactManager manager ("galera");
    QContactDetailFilter filter(QContactEmailAddress::match(email));
    QList<QContact> contacts = manager.contacts(filter);
    if(contacts.size() > 0) {
        qDebug() << "Result > 0";
        avatar = contacts[0].detail<QContactAvatar>().imageUrl().path();
    }

    QByteArray byteArray = avatar.toUtf8();
    char* cString = byteArray.data();

    callback(cString);
}

