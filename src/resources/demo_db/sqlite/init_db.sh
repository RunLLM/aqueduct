sqlite3 demo.db ".exit"

cat init_customer_activity.sql | sqlite3 demo.db
cat init_customers.sql | sqlite3 demo.db
cat init_hotel_reviews.sql | sqlite3 demo.db
cat init_mpg.sql | sqlite3 demo.db
cat init_tips.sql | sqlite3 demo.db
cat init_wine.sql | sqlite3 demo.db

sqlite3 demo.db <<EOF
.mode tabs
.import customer_activity.tsv customer_activity
EOF

sqlite3 demo.db <<EOF
.mode tabs
.import customers.tsv customers
EOF

sqlite3 demo.db <<EOF
.mode tabs
.import hotel_reviews.tsv hotel_reviews
EOF

sqlite3 demo.db <<EOF
.mode tabs
.import mpg.tsv mpg
EOF

sqlite3 demo.db <<EOF
.mode tabs
.import tips.tsv tips
EOF

sqlite3 demo.db <<EOF
.mode tabs
.import wine.tsv wine
EOF