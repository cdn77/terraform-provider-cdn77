$ terraform import cdn77_origin_object_storage.example <id>,<acl>,<cluster_id>

# <id> must be the ID (UUID) of the Object Storage Origin
# <acl> must be ACL type (one of: "authenticated-read", "private", "public-read", "public-read-write")
# <cluster_id> must be an ID (UUID) of the Object Storage cluster

# Example:
$ terraform import cdn77_origin_object_storage.example b2d6a7df-18df-4931-8c78-3842bc6e12f0,private,\
842b5641-b641-4723-ac81-f8cc286e288f
