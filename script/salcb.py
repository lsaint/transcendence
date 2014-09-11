# -*- coding: utf-8 -*-

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

