# -*- coding: utf-8 -*-
import time

class Timer(object):
    TimerDict = {}    # {tid:[time, fun, args]}
    Index = 0

    # 记录本次实例化后创建的timerid
    def __init__(self):
        self._TimerIdList = []


    # 释放本次实例化后的所有timer
    def ReleaseTimer(self):
        for idx in self._TimerIdList[:]:
            self.KillTimer(idx)
            self._TimerIdList.remove(idx)


    @classmethod
    def Update(cls):
        if len(Timer.TimerDict) == 0:
            return
        t = time.time()
        for tid, lt in Timer.TimerDict.items():
            if t - lt[0] >= lt[1]:
                lt[0] = t
                if not lt[2]:
                    cls.KillTimer(tid)
                apply(lt[3], lt[4])


    def setTimer(self, sec, is_loop, callbackfun, *args):
        Timer.Index += 1
        mark = time.time()
        Timer.TimerDict[Timer.Index] = [mark, sec, is_loop, callbackfun, args]
        self._TimerIdList.append(Timer.Index)

        return Timer.Index


    def SetTimer1(self, sec, cb, *args):
        return self.setTimer(sec, False, cb, *args)


    def SetTimer(self, sec, cb, *args):
        return self.setTimer(sec, True, cb, *args)


    def DoSetTimer(self, sec, cb, *args):
        cb(*args)
        return self.setTimer(sec, True, cb, *args)


    @classmethod
    def KillTimer(cls, tid):
        if Timer.TimerDict.has_key(tid):
            del Timer.TimerDict[tid]


g_timer = Timer()
