db.getCollection("posts").aggregate([
  {
    $match: {
      archived: false,
    },
  },
  {
    $lookup: {
      from: "users", // Collection to join with
      localField: "created_by", // Field from the Posts collection
      foreignField: "_id", // Field from the Users collection
      as: "userInfo", // Array field added to each post with user information
    },
  },
  {
    $unwind: "$userInfo", // Convert userInfo array to an object (assuming each post is created by one user)
  },
  {
    $project: {
      title: 1,
      description: 1,
      createdAt: 1,
      status: 1,
      created_at: 1,
      "userInfo.name": 1, // Only include the user's name in the output
    },
  },
  {
    $sort: {
      created_at: -1,
    },
  },
  {
    $facet: {
      metadata: [{ $count: "total" }, { $addFields: { page: 1 } }],
      data: [{ $skip: 0 }, { $limit: 10 }],
    },
  },
]);
