import re
import sys
import urllib.parse
from itertools import chain, combinations
import csv
import subprocess
import hashlib
from typing import BinaryIO
import logging

"""Take incoming HTTP requests and replay them with modified parameters."""
from mitmproxy import ctx, io, http

class Writer:
    def __init__(self, path: str) -> None:
        self.f: BinaryIO = open(path, "wb")
        self.w = io.FlowWriter(self.f)

    def done(self):
        self.f.close()

    def request( self, flow: http.HTTPFlow ) -> None:
        # Avoid an infinite loop by not replaying already replayed requests
        if flow.is_replay == "request":
            return
        flow_copy = flow.copy()
        url_for_payloads = search_for_url(flow_copy)
        if url_for_payloads:
            logging.warning(url_for_payloads)
            self.w.add(flow_copy)
            #go_subprocess = subprocess.Popen(['./ssrf-mutator',url_for_payloads['scheme'],url_for_payloads['netloc'],url_for_payloads['port'],url_for_payloads['path'],url_for_payloads['query']])
        # Only interactive tools have a view. If we have one, add a duplicate entry
        # for our flow_copy.
        if "view" in ctx.master.addons:
            ctx.master.commands.call("view.flow_copys.add", [flow_copy])
        
def search_for_url(flow):
    # Define the regular expression pattern for a URL
    url_pattern = r'(https?|ftp)://([^\s/:]+)(:[0-9]+)?(/[^?#]*)?(\?[^#]*)?(#.*)?'
    request_headers = str(flow.request.headers)
    request_content = str(flow.request.content)
    request_text = str(flow.request.text)
    raw_flow = str(flow)
    decoded_request = str(flow.request.decode())
    
    # Try to find the first match of the pattern within the input string
    match = re.search(url_pattern, request_content)
    
    if match:
        extracted_url = f"{match.string[match.start():].split(' ')[0]}"
        print(f"extracted_url: {extracted_url}")
        print(f"match.groups(): {match.groups()}")
       
        # Extract all parameters from the parameters string
        url_dict = dict()
        parsed_url_tuple = urllib.parse.urlparse(extracted_url)

        parsed_url_tuple_dict = parsed_url_tuple._asdict()
        for key in parsed_url_tuple_dict.keys():
            url_dict[key] = parsed_url_tuple_dict[key] 
        
        if parsed_url_tuple.scheme:
            url_dict['scheme'] = parsed_url_tuple.scheme 
        else:
            url_dict['scheme'] = ""
        if parsed_url_tuple.netloc:
            url_dict['netloc'] = parsed_url_tuple.netloc 
        else:
            url_dict['netloc'] = ""
        if parsed_url_tuple.username:
            url_dict['username'] = parsed_url_tuple.username 
        else:
            url_dict['username'] = ""
        if parsed_url_tuple.path:
            url_dict['path'] = parsed_url_tuple.path 
        else:
            url_dict['path'] = ""
        if parsed_url_tuple.params:
            url_dict['params'] = parsed_url_tuple.params 
        else:
            url_dict['params'] = ""
        if parsed_url_tuple.query:
            url_dict['query'] = parsed_url_tuple.query 
        else:
            url_dict['query'] = ""
        if parsed_url_tuple.fragment:
            url_dict['fragment'] = parsed_url_tuple.fragment 
        else:
            url_dict['fragment'] = ""
        if parsed_url_tuple.password:
            url_dict['password'] = parsed_url_tuple.password 
        else:
            url_dict['password'] = ""
        if parsed_url_tuple.hostname:
            url_dict['hostname'] = parsed_url_tuple.hostname 
        else:
            url_dict['hostname'] = ""
        if parsed_url_tuple.port:
            url_dict['port'] = parsed_url_tuple.port 
        else:
            url_dict['port'] = ""

        url_literal = f"{url_dict['scheme']}{url_dict['netloc']}{url_dict['port']}{url_dict['path']}{url_dict['query']}"

        param_pattern = r'([^&=]+)=([^&]*)'
        params = re.findall(param_pattern, url_dict['query'])
        params_dict = {}
        for p in params:
            params_dict[p[0].lstrip('?')] = p[1]
        url_dict['params_dict'] = params_dict
        
        filename_out = make_filename_from_path( url_dict['netloc'] , url_dict['path'] )
        with open(f"{filename_out}-raw.flow", "w") as outfile:
            outfile.write(request_content)
        content_to_file = re.sub(url_literal, "URLFUZZ", request_content)
        with open(f"{filename_out}-fuzzable.flow", "w") as outfile:
            outfile.write(content_to_file)

        return url_dict
    else:
        # If the input doesn't contain a URL, return None
        return None

def make_filename_from_path( netloc , path ):
    replaced_netloc = netloc.replace(".","-")
    replaced_path = hashlib.md5(path.replace("/","_").encode('utf-8')).hexdigest()
    return f"{replaced_netloc}_{replaced_path}"


addons = [Writer("test.writer")]

if __name__=='__main__':
    main()
