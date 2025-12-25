import argparse
import librosa
import numpy as np
import os

p = argparse.ArgumentParser()
p.add_argument("--input")
p.add_argument("--output")
p.add_argument("--segment")
args = p.parse_args()

y, sr = librosa.load(args.input, sr=16000, mono=True)

mel = librosa.feature.melspectrogram(
    y=y,
    sr=sr,
    n_mels=64,
    hop_length=512
)

os.makedirs(os.path.dirname(args.output), exist_ok=True)
np.save(args.output, mel)

print(f"Saved features for segment {args.segment}")
