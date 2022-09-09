# Get bearer token using username and password
curl -s -H "Content-Type: application/json" \
        -X POST -d "{\"username\": \"${DOCKER_LOGIN}\", \"password\": \"${DOCKER_PASSWORD}\"}" \
        https://hub.docker.com/v2/users/login/ | jq -r .token > token.jwt
export TOKEN=$(cat token.jwt)

# GET all repository tags
PAGE="https://hub.docker.com/v2/repositories/twdps/certificate-init-container/tags/"
while [[ "$PAGE" != null ]]; do
    curl ${PAGE} \
        -X GET \
        -H "Authorization: JWT ${TOKEN}" \
        > page_images.json
    cat page_images.json >> all_images.json
    PAGE=$(cat page_images.json | jq -r '.next')
done

# find development tagged images
cat all_images.json | jq -r '.results | .[] | select(.name | contains ("dev")) | .name' > dev_tags

# delete dev tagged images
while read TAG; do
    curl "https://hub.docker.com/v2/repositories/twdps/certificate-init-container/tags/${TAG}/" \
            -X DELETE \
            -H "Authorization: JWT ${TOKEN}"
done < dev_tags
