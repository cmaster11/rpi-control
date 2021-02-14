#!/usr/bin/python
# -*- coding: utf-8 -*-
import os
import select
import subprocess

import gi

gi.require_version("Gtk", "3.0")
from gi.repository import Gtk


# https://gist.github.com/mckaydavis/e96c1637d02bcf8a78e7
def run_bash_script(path):
    # create a pipe to receive stdout and stderr from process
    (pipe_r, pipe_w) = os.pipe()

    working_dir = '/home/pi'
    base_path = "/home/pi/code/rpi-control/scripts"
    full_path = base_path + path

    args = ['lxterminal', '--command', '/bin/bash ' + full_path]

    # lxterminal --command="/bin/bash -c '/home/pi/files.sh; read'"

    print('Executing ' + ' '.join(args))

    p = subprocess.Popen(args,
                         cwd=working_dir,
                         shell=False,
                         stdout=pipe_w,
                         stderr=pipe_w)

    # Loop while the process is executing
    while p.poll() is None:
        # Loop long as the select mechanism indicates there
        # is data to be read from the buffer
        while len(select.select([pipe_r], [], [], 0)[0]) == 1:
            # Read up to a 1 KB chunk of data
            buf = os.read(pipe_r, 1024)
            # Stream data to our stdout's fd of 0
            os.write(0, buf)

    # cleanup
    os.close(pipe_r)
    os.close(pipe_w)

    print(f'Command returned {p.returncode}')


def get_button(label, script_path):
    button = Gtk.Button.new_with_label(label)
    button.set_property("height-request", 60)

    def handler(_):
        run_bash_script(script_path)

    button.connect("clicked", handler)

    return button


class Window(Gtk.Window):

    def __init__(self):
        Gtk.Window.__init__(self, title="rPI Control")
        Gtk.Window.set_default_size(self, 400, 325)
        Gtk.Window.move(self, 0, 0)

        flowbox = Gtk.FlowBox()
        flowbox.set_valign(Gtk.Align.START)
        flowbox.set_max_children_per_line(4)
        flowbox.set_selection_mode(Gtk.SelectionMode.NONE)

        # --- Buttons

        flowbox.add(get_button('cmaster11-HURR On!', '/wol-cmaster11-hurr.sh'))

        table_box = Gtk.Box(spacing=6)
        table_box.add(get_button('Table Up!', '/table-up.sh'))
        table_box.add(get_button('Table Down!', '/table-down.sh'))

        flowbox.add(table_box)

        # --- END Buttons

        self.add(flowbox)
        self.show_all()


Gtk.init()

window = Window()
window.connect("delete-event", Gtk.main_quit)
window.show_all()
Gtk.main()
