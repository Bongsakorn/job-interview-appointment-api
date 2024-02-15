db.createUser(
    {
        user: "recruiter",
        pwd: "Zxsw2#edcv",
        roles: [
            {
                role: "readWrite",
                db: "recruiting"
            }
        ]
    }
);

db.getCollection('users').createIndex(
    {
        "email": 1,
        "password": 1,
    },
    {
        "unique": true,
        "sparse": true,
        "background": true,
        "name": "login_idx"
    }
);

db.getCollection('users').insertOne(
    {
      "email": "user1@example.com",
      "password": "$2a$14$PrKCOlqAW7GGiP5OUI8uk.k2YiQM6mrMRmgYcyk7Qs2VPTdh2aKa.",
      "name": "ยูสเซอร์หนึ่ง",
      "avatar_url": "",
      "created_at": new Date(),
    }
);