if [ ! -d ~/tests ]; then
  mkdir -p ~/tests;
fi
sqlite3 ~/tests/test.db "VACUUM;"