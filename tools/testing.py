# -*- coding: utf8 -*-

import salcb
from logic_pb2 import *

def simulateSalClientProto():
    print "test-simulateSalClientProto"
    pb = C2SLogin()
    pb.uid = 900
    salcb.OnSALClientProto(1640285, 50001906, 101, pb.SerializeToString())


def simulateRecvClientProto(tsid, ssid, uri, data, action, uids):
    if uri == 101:
        ins = C2SLogin()
        ins.ParseFromString(data)
        print "recv proto:", ins
    else:
        print "expect uri101, got:", uri
