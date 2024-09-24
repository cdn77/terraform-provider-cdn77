$ terraform import cdn77_origin_aws.example <id>,<access_key_secret>

# <id> must be the ID (UUID) of the AWS Origin
# <access_key_secret> must be the secret of the AWS access key that you provided when creating the AWS Origin

# Example:
$ terraform import cdn77_origin_aws.example 4cd2378b-dec8-49e2-aa17-bf7561452998,VWK92izmd7zpY8Khs/Dllv4yLYc4sFWNyg2XtuNF
