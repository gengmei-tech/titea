#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2018 yongman <yming0221@gmail.com>
#
# Distributed under terms of the MIT license.

"""
redis client wrapper
"""

import redis

ip = "127.0.0.1"
port = 5379

class RedisWrapper:
    def __init__(self):
        self.r = redis.StrictRedis(host=ip, port=port)

    def get_instance(self):
        return self.r
