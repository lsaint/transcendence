# -*- coding: utf-8 -*-

import sys; sys.path.extend(["./conf/", "./proto/", "./script/"])
import log
import traceback
import server_pb2, logic_pb2
import go
import post

from timer import Timer
from config import *




def OnProto(tsid, ssid, uri, data, uid):
    try:
        log.debug("OnProto--> tsid:%s ssid:%s uri:%d len:%d" % (tsid, ssid, uri, len(data)))
    except Exception as err:
        log.error("%s-%s" % ("OnProto", traceback.format_exc()))


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

# sal

def SALSubscribeUserInOutMove(svctype, ):
    pass

# 

def test():
    log.debug("testtestbanbang")
