import argparse
import time
import sys

p = argparse.ArgumentParser()
p.add_argument("--job-id")
p.add_argument("--dataset")
p.add_argument("--model")
args = p.parse_args()

print("Training", args.job_id)
time.sleep(5)

sys.exit(0)
