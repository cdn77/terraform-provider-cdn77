$ terraform import cdn77_origin.example <id>,<type>[,type-specific parameters,...]

# <id> must be the ID (UUID) of the Origin
# <type> must be a type of the Origin (one of: "aws", "object-storage", "url")
# Depending on the type of the origin there may be other required parameters:

# URL type doesn't need other parameters
$ terraform import cdn77_origin.example <id>,url

# AWS type requires access key secret
# <aws_access_key_secret> must be the secret of the AWS access key that you provided when creating the Origin
$ terraform import cdn77_origin.example <id>,aws,<aws_access_key_secret>

# Object Storage type requires ACL, cluster ID, access key ID and access key secret
# <acl> must be ACL type (one of: "authenticated-read", "private", "public-read", "public-read-write")
# <cluster_id> must be an ID (UUID) of the Object Storage cluster
# <access_key_id> must be the ID of the access key that was returned when creating the Origin
# <access_key_secret> must be the secret of the access key that was returned when creating the Origin
$ terraform import cdn77_origin.example <id>,object-storage,<acl>,<cluster_id>,<access_key_id>,<access_key_secret>

# Examples:
$ terraform import cdn77_origin.example_url 4cd2378b-dec8-49e2-aa17-bf7561452998,url
$ terraform import cdn77_origin.example_aws 4cd2378b-dec8-49e2-aa17-bf7561452998,aws,\
VWK92izmd7zpY8Khs/Dllv4yLYc4sFWNyg2XtuNF
$ terraform import cdn77_origin.example_object_storage 4cd2378b-dec8-49e2-aa17-bf7561452998,object-storage,\
private,842b5641-b641-4723-ac81-f8cc286e288f,I17DXFE00GNJZVQUTQPW,7UG7WbcIz4VhZnVxV4XQcDR2X0APApuvthyATf2v
