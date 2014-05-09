package main

type CanopyCass struct {
    cluster *gocql.ClusterConfig
    session *gocql.Session
}

func CanopyCass_Connect() CanopyCass {
    cluster := gocql.NewCluster("127.0.0.1")
    cluster.Keyspace = "example"
    cluster.Consistency = gocql.Quorum;
    session, _ := cluster.CreateSession();
    return cass;
}
