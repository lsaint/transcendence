# -*- coding: utf-8 -*-

import sys; sys.path.extend(["./conf/", "./proto/", "./script/"])
import log
import traceback
import server_pb2, logic_pb2
import go
import post

from timer import Timer
from config import *




def OnGateProto(tsid, ssid, uri, data, action, uids):
    try:
        log.debug("OnGateProto--> tsid:%s ssid:%s uri:%d len:%d" % (tsid, ssid, uri, len(data)))

        # hive recv client proto
        if action == server_pb2.Recv:
            return

        # drone do cast
        if action == server_pb2.Broadcast:
            return
        elif action == server_pb2.Unicast:
            return
        elif action == server_pb2.Multicast:
            return

    except Exception as err:
        log.error("%s-%s" % ("OnGateProto", traceback.format_exc()))


def OnTicker():
    try:
        Timer.Update()
    except Exception as err:
        log.error("%s-%s" % ("OnTicker", traceback.format_exc()))


def OnPostDone(sn, ret):
    try:
        post.OnPostDone(sn, ret)
    except Exception as err:
        log.error("%s-%s" % ("OnPostDone", traceback.format_exc()))


def OnHttpReq(jn, url):
    log.debug("OnHttpReq--> json: %s, url: %s" % (jn, url))
    return "return " + url


def test():
    log.debug("testtestbanbang")
