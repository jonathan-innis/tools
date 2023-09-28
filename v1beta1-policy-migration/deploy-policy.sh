export AWS_PARTITION=aws
export AWS_REGION=us-west-2
export AWS_ACCOUNT_ID=330700974597
export CLUSTER_NAME=scale-test

POLICY_DOCUMENT=$(envsubst < policy_document.json)
POLICY_NAME="KarpenterControllerPolicy-$CLUSTER_NAME-v1beta1"
ROLE_NAME="$CLUSTER_NAME-karpenter"

echo "Creating policy $POLICY_NAME..."
POLICY_ARN=$(aws iam create-policy --policy-name "$POLICY_NAME" --policy-document "$POLICY_DOCUMENT" | jq -r .Policy.Arn)
echo "Attaching policy to role $ROLE_NAME..."
aws iam attach-role-policy --role-name "$ROLE_NAME" --policy-arn "$POLICY_ARN"
echo "Finished attaching Karpenter policy"