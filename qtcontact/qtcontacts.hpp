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

#ifndef __QTCONTACTS_HPP_
#define __QTCONTACTS_HPP_

#include <QObject>
#include <QSignalMapper>

class Avatar : QObject {
    Q_OBJECT
    public:
        explicit Avatar(QObject* parent=0)
            : QObject(parent) {
            _signalMapper = new QSignalMapper(this);
            // connect empty signal
            connect(this, SIGNAL(readyToRetrieve()),
                _signalMapper, SLOT(map()));
            // connect signal that takes string
            connect(_signalMapper, SIGNAL(mapped(QString)),
                this, SLOT(retrieveThumbnail(QString)));
        }

        ~Avatar() {
            _signalMapper->deleteLater();
        }

        void getThumbnail(char *email);

    public slots:
        void emitSignals();
        void retrieveThumbnail(const QString& email);

    signals:
        void readyToRetrieve();

    private:
        QSignalMapper* _signalMapper = nullptr;
};

#endif
