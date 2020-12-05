import argparse
import json
from types import SimpleNamespace
import sr
import logging


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('config_path', help='path to config json file')
    args = parser.parse_args()

    with open(args.config_path) as f:
        s = f.read()
        d = json.loads(s, object_hook=lambda d: SimpleNamespace(**d))
    sr_list = sr.SRList(sid=d.list.sid, dev=d.list.dev)
    sr_policy = sr.Policy(source_addr=d.source, dest_addr=d.dest, sr_list=sr_list)
    sr_policy.set_encap()
    logging.info("seg6 encap succeeded")
