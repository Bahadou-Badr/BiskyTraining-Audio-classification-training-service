import argparse
import json
import os
import time
import sys

parser = argparse.ArgumentParser()
parser.add_argument("--job-id", required=True)
parser.add_argument("--dataset", required=True)
parser.add_argument("--model", required=True)
parser.add_argument("--out", default="artifacts")
args = parser.parse_args()

job_id = args.job_id
out_dir = os.path.join(args.out, job_id)
os.makedirs(out_dir, exist_ok=True)

# Simulate training
print(f"Training job {job_id} on dataset {args.dataset}", file=sys.stderr)
time.sleep(5)

# Fake model artifact
model_path = os.path.join(out_dir, "model.bin")
with open(model_path, "wb") as f:
    f.write(b"FAKE_MODEL_BINARY")

# Fake metrics
metrics = {
    "accuracy": 0.91,
    "loss": 0.08
}

params = {
    "epochs": 10,
    "lr": 0.001
}

result = {
    "metrics": metrics,
    "params": params,
    "artifact_path": model_path
}

print(json.dumps(result))
sys.exit(0)
