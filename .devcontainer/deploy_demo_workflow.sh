aqueduct start --disable-usage-stats &
sleep 30
python3 /deploy_notebooks/initialize.py --example-notebooks-only --wait-to-complete