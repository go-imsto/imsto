#!/sbin/runscript
# Copyright 1999-2013 Gentoo Foundation
# Distributed under the terms of the GNU General Public License v2

# copy this file into /etc/init.d/ and named 'imsto'

IMSTO_APP_ROOT=${IMSTO_APP_ROOT:-/opt/imsto}
IMSTO_API_0_SALT=${IMSTO_API_0_SALT:-imstosalt}
ACT=${SVCNAME#*.}
if [ -z "${ACT}" -o ${ACT} != "stage" ]; then
	ACT="tiring"
fi

IM_PID="/var/run/imsto.${ACT}.pid"
# IM_LOG="/var/log/imsto/${ACT}_log"
IM_CONF="${IMSTO_APP_ROOT}/config/imsto.ini"

IM_BIN="/usr/local/bin/imsto"

depend() {
	need localmount net
	before nginx
}

start_pre() {
	[ -d /var/run/imsto ] || mkdir /var/run/imsto && chown nobody /var/run/imsto
	[ -d /var/log/imsto ] || mkdir /var/log/imsto && chown nobody /var/log/imsto
}

start() {
	ebegin "Starting imsto.${ACT}"
	start-stop-daemon --background --start --exec \
	env IMSTO_API_0_SALT=${IMSTO_API_0_SALT} \
	${IM_BIN} \
	-u nobody \
	--make-pidfile --pidfile ${IM_PID} \
	-- -root="${IMSTO_APP_ROOT}" -logs="/var/log/imsto" ${ACT}
	eend $?
}

stop() {
    ebegin "Stopping imsto.${ACT}"
    start-stop-daemon --stop --exec ${IM_BIN} \
    --pidfile ${IM_PID}
    eend $?
}
