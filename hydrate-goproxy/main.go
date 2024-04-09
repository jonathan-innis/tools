package main

import (
	"github.com/aws/karpenter-provider-aws/pkg/operator"

	coreoperator "sigs.k8s.io/karpenter/pkg/operator"
)

func main() {
	_, _ = operator.NewOperator(coreoperator.NewOperator())
}
