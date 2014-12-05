# -*- coding: utf8 -*-

import base64, random
#import sal, salcb
import go
from timer import g_timer, g_leadertimer
from logic_pb2 import *
from server_pb2 import *


def subscribeSal():
    g_timer.SetTimer(30, sal.SALPing)
    print "SALSocketInfo", sal.SALSocketInfo("127.0.0.1", 9090)
    print "SALSSubscribeUserInOutMove", sal.SALSSubscribeUserInOutMove(1640285)


def simulateSalClientProto():
    print "test-simulateSalClientProto"
    pb = C2SLogin()
    pb.uid = 900
    salcb.OnSALClientProto(1640285, 50001906, 101, pb.SerializeToString())


def simulateRecvClientProto(tsid, ssid, uri, data, action, uids):
    if uri == 101:
        ins = C2SLogin()
        ins.ParseFromString(base64.b64decode(data))
        print "recv proto:", ins.DESCRIPTOR.name, ins
    else:
        print "got proto uri =", uri


def simulateSendProtoToClient():
    pb = S2CLoginRep()
    pb.role = "administer"
    go.SendMsg(1640285, 1640285, 102, pb.SerializeToString(), H2D_Unicast, 0, 50001906)


def simulateDroneRecvUnicast(tsid, ssid, uri, data, action, uids):
    if uri == 102:
        pb = S2CLoginRep()
        pb.ParseFromString(base64.b64decode(data))
        print "recv proto:", pb.DESCRIPTOR.name, pb
    else:
        print "got proto uri =", uri


def testIsLeader():
    def lam():
        print(go.IsLeader())
    g_timer.SetTimer(5, lam)



def func():
    print("leader timer tick tack. tick tack..")
def testLeaderTimer():
    g_leadertimer.SetLeaderTimer(5, func)



import raft
def testRaftApply():
    raft.apply("hello raft %d" % (random.randint(1, 1000)))


import select, os, struct, sys
def testEpoll(fd):
    epoll = select.epoll()
    epoll.register(fd, select.EPOLLIN)
    while True:
        events = epoll.poll(1)
        for fileno, event in events:
            task = go.GetTask()
            ret = os.read(fd, 8)
            print "read fd", struct.unpack("Q", ret)[0]
            # do task
            # clear task


def testKqueue():
    import signal
    kq = select.kqueue()
    kevent = select.kevent(signal.SIGUSR1, filter=select.KQ_FILTER_SIGNAL,
                            flags=select.KQ_EV_ADD | select.KQ_EV_ENABLE)
    while True:
        revents = kq.control([kevent], 1, None)  # block
        for event in revents:
            if (event.filter == select.KQ_FILTER_SIGNAL):
                go.DoTask()


def testEventNotify(fd):
    import platform
    if platform.system() == "Linux":
        testEpoll(fd)
    else:
        testKqueue()

