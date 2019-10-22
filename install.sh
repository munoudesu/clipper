#!/bin/bash

INSTALLDIR="/usr/local/clipper"
if [ -n "$1" ]; then
	INSTALLDIR=$1
fi	
if [ -x clipper ]; then
	install -o root -g root -m 755 -D  clipper $INSTALLDIR/bin/clipper
elif [ -x $GOPATH/bin/clipper ]; then
	install -o root -g root -m 755 -D  $GOPATH/bin/clipper $INSTALLDIR/bin/clipper
fi
install -o root -g root -m 644 -D template/index.tmpl            $INSTALLDIR/source/template/index.tmpl
install -o root -g root -m 644 -D template/page.tmpl             $INSTALLDIR/source/template/page.tmpl
install -o root -g root -m 644 -D resource/root/css/clipper.css  $INSTALLDIR/source/resource/root/css/clipper.css
install -o root -g root -m 644 -D resource/root/icon/favicon.ico $INSTALLDIR/source/resource/root/icon/favicon.ico
install -o root -g root -m 644 -D resource/root/js/axios.min.js  $INSTALLDIR/source/resource/root/js/axios.min.js
install -o root -g root -m 644 -D resource/root/js/axios.min.map $INSTALLDIR/source/resource/root/js/axios.min.map
install -o root -g root -m 644 -D resource/root/js/clipper.js    $INSTALLDIR/source/resource/root/js/clipper.js
install -o root -g root -m 644 -D resource/root/js/vue.min.js    $INSTALLDIR/source/resource/root/js/vue.min.js
install -o root -g root -m 644 -D clipper.conf                   $INSTALLDIR/etc/clipper.conf
install -o root -g root -m 644 -D clipper.service                $INSTALLDIR/etc/systemd/system/clipper.service
install -o root -g root -m 644 -D clipper.cron                   $INSTALLDIR/etc/cron.d/clipper
ln -sf $INSTALLDIR/etc/systemd/system/clipper.service /lib/systemd/system/clipper.service
ln -sf $INSTALLDIR/etc/cron.d/clipper /etc/cron.d/clipper
mkdir -p $INSTALLDIR/etc/default
chmod 755 $INSTALLDIR/etc/default
chown root:root $INSTALLDIR/etc/default
mkdir -p $INSTALLDIR/db
chmod 755 $INSTALLDIR/db
chown root:root $INSTALLDIR/db
if [ -f clipper.db ]; then
	install -o root -g root -m 644 -D clipper.db $INSTALLDIR/db/clipper.db
fi
mkdir -p $INSTALLDIR/build
chmod 755 $INSTALLDIR/build
chown root:root $INSTALLDIR/build
touch $INSTALLDIR/etc/youtube_data_api_key_file
chmod 600 $INSTALLDIR/etc/youtube_data_api_key_file
chown root:root $INSTALLDIR/etc/youtube_data_api_key_file
if [ -f youtube_data_api_key_file ]; then
	install -o root -g root -m 600 -D youtube_data_api_key_file $INSTALLDIR/etc/youtube_data_api_key_file
fi
touch $INSTALLDIR/etc/twitter_api_key_file
chmod 600 $INSTALLDIR/etc/twitter_api_key_file
chown root:root $INSTALLDIR/etc/twitter_api_key_file
if [ -f twitter_api_key_file ]; then
	install -o root -g root -m 600 -D twitter_api_key_file $INSTALLDIR/etc/twitter_api_key_file
fi
