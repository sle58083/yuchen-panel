#!/usr/bin/env python3
# Synthetic digest fixtures are for unit tests only, never main/version.json.
import importlib.util
import unittest
from pathlib import Path
spec=importlib.util.spec_from_file_location("release_manifest",Path(__file__).with_name("release-manifest.py"))
module=importlib.util.module_from_spec(spec); spec.loader.exec_module(module)
class ReleaseMetadataTest(unittest.TestCase):
    def test_rc_fields_are_consistent(self):
        data=module.release_manifest("1.0.3-rc.1", "a"*64, "sle58083/yuchen-panel")
        self.assertEqual(data["latest"],data["version"])
        self.assertEqual(data["version"],"1.0.3-rc.1")
        self.assertEqual(data["codename"],"rc")
        self.assertEqual(data["package"],"yuchen-panel-v1.0.3-rc.1.zip")
        self.assertIn("/v1.0.3-rc.1/",data["download_url"])
        self.assertNotIn("v1.0.2",data["download_url"])
    def test_rejects_bad_digest(self):
        with self.assertRaises(ValueError): module.release_manifest("1.0.3-rc.1","invalid","sle58083/yuchen-panel")
if __name__=="__main__": unittest.main(verbosity=2)
