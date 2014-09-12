# -*- coding: utf-8 -*-

import sys; sys.path.extend(["./conf/", "./proto/", "./script/"])
import go
import server_pb2

def SALSubscribeUserInOutMove(tp_inout):
    print "py SALSubscribeUserInOutMove", tp_inout


def SALSubscribeMaixuQueueChange(subscribe_mx):
    print "py SALSubscribeMaixuQueueChange", subscribe_mx


def SALQueryMaixuQueue(maixu_queue):
    print "py SALQueryMaixuQueue", maixu_queue


def SALQueryUserRole(tp_user_role):
    print  "py SALQueryUserRole", tp_user_role


# drone 
# data是uri所表示的完整协议数据
def OnSALClientProto(tsid, uid, uri, data):
    print "py OnSALClientProto", tsid, uid, uri, len(data)
    #go.SendMsg(tsid, ssid, uri, data, server_pb2.D2H_Msg, 0, uid)
