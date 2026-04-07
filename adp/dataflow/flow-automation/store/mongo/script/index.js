// "dag_instance" should replace with your collection name
db.dag_instance.createIndex(
    {
        "cmd": 1
    },
    {
        name: "cmd_index",
    }
);
db.dag_instance.createIndex(
    {
        "status": 1
    },
    {
        name: "status_index",
    }
);
db.dag_instance.createIndex(
    {
        "updated_at": 1
    },
    {
        name: "updated_at_index",
    }
);

// "task_instance" should replace with your collection name
db.task_instance.createIndex(
    {
        "status": 1
    },
    {
        name: "status_index",
    }
);
db.task_instance.createIndex(
    {
        "dagInsId": 1
    },
    {
        name: "dag_ins_id_index",
    }
);
db.task_instance.createIndex(
    {
        "updated_at": 1
    },
    {
        name: "updated_at_index",
    }
);