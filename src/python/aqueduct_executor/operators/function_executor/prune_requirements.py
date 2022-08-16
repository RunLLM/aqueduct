import argparse


def run(local_path: str, requirements_path: str, missing_path: str) -> None:
    with open(local_path, "r") as f:
        local_req = set(f.read().split("\n"))

    with open(requirements_path, "r") as f:
        required = f.read().split("\n")

    missing = []
    for r in required:
        # Remove any @ file because we may not have those files local to the user's device in our file system.
        if r not in local_req and "@ file" not in r:
            missing.append(r)

    if len(missing) > 0:
        with open(missing_path, "w") as f:
            f.write("\n".join(missing))


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--local_path", required=True)
    parser.add_argument("--requirements_path", required=True)
    parser.add_argument("--missing_path", required=True)
    args = parser.parse_args()

    run(args.local_path, args.requirements_path, args.missing_path)
