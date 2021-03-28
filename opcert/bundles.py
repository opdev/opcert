#! /usr/bin/env python3

import yaml
import subprocess
import os
import shutil
import ipywidgets as widgets
from IPython.display import display

class Bundles:
    def __init__(self):
        # loding input path
        self.bundles = []
   
    def Clean(self):
         [shutil.rmtree(d) for d in os.listdir('.') if d.startswith('manifests-')]

    def Download(self):
        cmd = ["offline-cataloger", "generate-manifests", "certified-operators"]
        process = subprocess.Popen(cmd)
        output = process.communicate()[0]
        process.wait()
        [[self.bundles.append(subd) for subd in os.listdir(d)] for d in os.listdir('.') if d.startswith('manifests-')]

    def List(self):
        self.Update()
        w = widgets.Dropdown(
                options=self.bundles,
                description='',
                disabled=False,
            )
        display(w)

    def Update(self):
        self.Clean()
        self.Download()

