aqueduct start --disable-usage-stats &
sleep 30
python3 /deploy_notebooks/initialize.py --demo-container-notebooks-only --wait-to-complete