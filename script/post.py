# -*- coding: utf-8 -*-

import go, log

g_postsn = 0
g_post_callback = {}


def GetPostSn():
    global g_postsn
    g_postsn += 1
    return g_postsn


def OnPostDone(sn, ret):
    global g_post_callback
    cb = g_post_callback.get(sn)
    if cb:
        cb(sn, ret)
        del g_post_callback[sn]
    else:
        log.warn("not exist post sn %d" % sn)


def PostAsync(url, s, func=None, sn=None):
    global g_post_callback
    sn = sn or GetPostSn()
    go.PostAsync(url, s, sn)
    if func:
        g_post_callback[sn] = func
    return sn
