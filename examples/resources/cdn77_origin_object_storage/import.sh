$ terraform import cdn77_origin_object_storage.example <id>,<acl>,<cluster_id>,<access_key_id>,<access_key_secret>

# <id> must be the ID (UUID) of the Object Storage Origin
# <acl> must be ACL type (one of: "authenticated-read", "private", "public-read", "public-read-write")
# <cluster_id> must be an ID (UUID) of the Object Storage cluster
# <access_key_id> must be the ID of the access key that was returned when creating the Object Storage Origin
# <access_key_secret> must be the secret of the access key that was returned when creating the Object Storage Origin

# Example:
$ terraform import cdn77_origin_object_storage.example b2d6a7df-18df-4931-8c78-3842bc6e12f0,private,\
842b5641-b641-4723-ac81-f8cc286e288f,I17DXFE00GNJZVQUTQPW,7UG7WbcIz4VhZnVxV4XQcDR2X0APApuvthyATf2v
