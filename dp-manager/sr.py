from typing import List
from pyroute2 import IPRoute, NDB


class SRList:
    def __init__(self, sid: List[str], dev: str):
        self.sid = sid
        self.dev = dev


class Policy:
    def __init__(self, source_addr: str, dest_addr: str, sr_list: SRList):
        self.source_addr = source_addr
        self.dest_addr = dest_addr
        self.sr_list = sr_list

    def set_encap(self):
        ip = IPRoute()
        segs = ",".join(self.sr_list.sid)
        ifidx = self.get_ifidx_byname()
        ip.route('add',
                 dst=self.dest_addr,
                 oif=ifidx,
                 encap={'type': 'seg6',
                        'mode': 'encap',
                        'segs': segs}
                 )

    def get_ifidx_byname(self):
        with NDB(log='on') as ndb:
            ifidx = ndb.interfaces[self.sr_list.dev]["index"]
        return ifidx
