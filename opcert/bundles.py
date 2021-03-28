#! /usr/bin/env python3

import yaml
import subprocess
import os
import shutil

class Bundles:
    def __init__(self):
        # loding input path
        self.bundles = []
   
    def Clean(self):
         [shutil.rmtree(d) for d in os.listdir('.') if d.startswith('manifests-')]

    def Download(self):
        print("\n\nDownloading manifests for selected certified operators... \n\n")
        cmd = ["offline-cataloger", "generate-manifests", "certified-operators"]
        process = subprocess.Popen(cmd)
        output = process.communicate()[0]
        process.wait()
        print("Done.")

    def Update(self):
        self.Clean()
        self.Download()
        
    def List(self):
        print("\n\nListing existing certified bundles...\n")
        print("---------------------------------")
        [[print(subd) for subd in os.listdir(d)] for d in os.listdir('.') if d.startswith('manifests-')]