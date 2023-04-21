import argparse
from huggingface_hub import snapshot_download

def main(args):
    snapshot_download(
        repo_id=args.repo_id,
        local_dir=args.local_dir,
        local_dir_use_symlinks=False,
    )

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo-id", type=str)
    parser.add_argument("--local-dir", type=str)
    args = parser.parse_args()

    main(args)