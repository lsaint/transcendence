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

def SALSubscribeUserInOutMove(tp_inout):
    print "py SALSubscribeUserInOutMove", tp_inout


def SALSubscribeMaixuQueueChange(subscribe_mx):
    print "py SALSubscribeMaixuQueueChange", subscribe_mx


def SALQueryMaixuQueue(maixu_queue):
    print "py SALQueryMaixuQueue", maixu_queue


def SALQueryUserRole(tp_user_role):
    print  "py SALQueryUserRole", tp_user_role


#def SALMsgFromClient(msg):
#    print "py SALMsgFromClient", msg

# data是uri所表示的完整协议数据
def OnSALClientProto(tsid, uid, uri, data):
    print "py OnSALClientProto", tsid, uid, uri, data

# 

def test():
    log.debug("testtestbanbang")
