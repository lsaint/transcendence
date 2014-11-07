# -*- coding: utf8 -*-

import base64
import sal, salcb, go
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

