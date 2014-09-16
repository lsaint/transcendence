# -*- coding: utf-8 -*-

import sys; sys.path.extend(["./conf/", "./proto/", "./script/", "./tools/"])
import traceback, struct
import server_pb2, logic_pb2
import log, go, post, sal

from timer import Timer
from config import *




def OnGateProto(tsid, ssid, uri, data, action, uids):
    try:
        log.debug("OnGateProto--> tsid:%s ssid:%s uri:%d len:%d" % (tsid, ssid, uri, len(data)))

        import testing

        # hive recv client proto 
        # trigger app logic
        # then go.SendMsg to client
        if action == server_pb2.D2H_Msg:
            testing.simulateRecvClientProto(tsid, ssid, uri, data, action, uids)

        # drone connected to hive 
        elif action == server_pb2.D2H_Register:
            pass

        # drone cast sal
        elif action == server_pb2.H2D_Broadcast:
            testing.simulateDroneRecvUnicast(tsid, ssid, uri, data, action, uids):
            sal.SALSubSidBroadcast(tsid, ssid, 0, packProto(uri, data))
        elif action == server_pb2.H2D_Unicast:
            sal.SALUnicast(tsid, uids[0], packProto(uri, data))
        elif action == server_pb2.H2D_Multicast:
            sal.SALMulticast2(tsid, 0, packProto, uids)

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
    # test
    import testing
    if url == "/transcendence1":
        testing.simulateSalClientProto()
    else:
        testing.simulateSendProtoToClient()
    return "return " + url


def test_script():
    log.debug("testtestbanbang")


def packProto(uri, data):
    return "%s%s" % (struct.pack("II", len(data) + 8, uri), data)
