# -*- coding: utf-8 -*-

import sys; sys.path.extend(["./conf/", "./proto/", "./script/", "./tools/"])
import traceback, struct
import server_pb2, logic_pb2
import log, go, post

from timer import Timer, LeaderTimer
from config import *
from const import *




def OnGateProto(tsid, ssid, uri, data, action, uids):
    try:
        log.debug("OnGateProto--> tsid:%s ssid:%s uri:%d len:%d action: %s" % (
                    tsid, ssid, uri, len(data), action))

        import testing

        # hive recv client proto 
        # trigger app logic
        # then go.SendMsg to client
        if action == server_pb2.D2H_Msg:
            testing.simulateRecvClientProto(tsid, ssid, uri, data, action, uids)

        # drone connected to hive 
        # hive init
        elif action == server_pb2.D2H_Register:
            testing.subscribeSal()

        # drone cast sal
        elif action == server_pb2.H2D_Broadcast:
            sal.SALSubSidBroadcast(tsid, ssid, 0, packProto(uri, data))
        elif action == server_pb2.H2D_Unicast:
            testing.simulateDroneRecvUnicast(tsid, ssid, uri, data, action, uids)
            sal.SALUnicast(tsid, uids[0], packProto(uri, data))
        elif action == server_pb2.H2D_Multicast:
            sal.SALMulticast2(tsid, 0, packProto, uids)

    except Exception as err:
        log.error("%s-%s" % ("OnGateProto", traceback.format_exc()))


def OnTicker():
    try:
        Timer.Update()
        LeaderTimer.Update()
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
    testing.testRaftApply()

    #if url == "/transcendence1":
    #    testing.simulateSalClientProto()
    #else:
    #    testing.simulateSendProtoToClient()

    return "return " + url


def OnUplinkmsg(meta, uri, data):
    print "OnUplinkmsg", meta, uri, data


def OnLeaveplatform(meta, uid):
    print "OnLeaveplatform", meta, uid



def OnClusterNodeEvent(ev_type, node_name):
    print "----->OnClusterNodeEvent<----", ev_type, node_name
    try:
        # become leader
        if ev_type == EV_NODE_BE_LEADER:
            LeaderTimer.OnBecomeLeader()
            #import testing
            #testing.testLeaderTimer()
        elif ev_type == EV_NODE_OFF_LEADER:
            LeaderTimer.OnHandoffLeader()
    except Exception as err:
        log.error("%s-%s" % ("OnClusterNodeEvent", traceback.format_exc()))


def OnRaftApply(data):
    print "OnRaftApply:", data


def test_script():
    log.debug("testtestbanbang")


def packProto(uri, data):
    return "%s%s" % (struct.pack("II", len(data) + 8, uri), data)


import thread, time
def main(fd):
    print "#*-> main <-*#", fd
    try:
        thread.start_new_thread(pymain, ())
        loop(fd)
    except Exception as err:
        log.error("%s-%s" % ("main", traceback.format_exc()))


def pymain():
    print "pymain start"
    print "pymain end"


import select, os, struct, sys
def EpollLoop(fd):
    epoll = select.epoll()
    epoll.register(fd, select.EPOLLIN)
    go.PyReady()
    while True:
        events = epoll.poll(1)
        for fileno, event in events:
            ret = os.read(fd, 8)
            go.DoTask()


def KqueueLoop():
    import signal
    kq = select.kqueue()
    kevent = select.kevent(signal.SIGUSR1, filter=select.KQ_FILTER_SIGNAL,
                            flags=select.KQ_EV_ADD | select.KQ_EV_ENABLE)
    go.PyReady()
    while True:
        revents = kq.control([kevent], 1, None)  # block
        for event in revents:
            if (event.filter == select.KQ_FILTER_SIGNAL):
                go.DoTask()


def loop(fd):
    import platform
    if platform.system() == "Linux":
        EpollLoop(fd)
    else:
        KqueueLoop()


