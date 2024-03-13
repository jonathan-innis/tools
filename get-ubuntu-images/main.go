package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/samber/lo"
)

type Options struct {
	Regions []string
	Images  []string
	OutPath string
}

func ParseFlags() (*Options, error) {
	regionsString := flag.String("regions", "", "A comma separated list of regions to query. If unspecified, queries all commercial regions.")
	imagesString := flag.String("images", "", "A comma separated list of image names to query.")
	outputPath := flag.String("out", "ami_ids.csv", "File to write the results.")
	flag.Parse()

	opts := &Options{}
	opts.Regions = lo.Ternary(lo.FromPtr(regionsString) == "", []string{
		"af-south-1",
		"ap-east-1",
		"ap-northeast-1",
		"ap-northeast-2",
		"ap-northeast-3",
		"ap-south-1",
		"ap-south-2",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-southeast-3",
		"ca-central-1",
		"ca-west-1",
		"eu-central-1",
		"eu-central-2",
		"eu-north-1",
		"eu-south-1",
		"eu-south-2",
		"eu-west-1",
		"eu-west-2",
		"eu-west-3",
		"il-central-1",
		"me-central-1",
		"me-south-1",
		"sa-east-1",
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
	}, strings.Split(lo.FromPtr(regionsString), ","))
	if lo.FromPtr(imagesString) == "" {
		return nil, fmt.Errorf("must specify --images")
	}
	opts.Images = strings.Split(lo.FromPtr(imagesString), ",")
	opts.OutPath = lo.FromPtr(outputPath)
	return opts, nil
}

type AMIEntry struct {
	Name   string
	ID     string
	Region string
}

func main() {
	opts, err := ParseFlags()
	if err != nil {
		panic(fmt.Errorf("parsing flags, %w", err))
	}

	ctx := context.Background()
	cfg := lo.Must(config.LoadDefaultConfig(ctx))
	ec2Client := ec2.NewFromConfig(cfg)

	file := lo.Must(os.OpenFile(opts.OutPath, os.O_RDWR|os.O_CREATE, 0777))
	defer file.Close()
	w := csv.NewWriter(file)
	defer w.Flush()


	failedRegions := []string{}
	entries := []AMIEntry{}
	for _, region := range opts.Regions {
		out, err := ec2Client.DescribeImages(ctx, &ec2.DescribeImagesInput{
			Filters: []types.Filter{
				{
					Name:   lo.ToPtr("name"),
					Values: opts.Images,
				},
			},
		}, WithRegion(region))
		if err != nil {
			failedRegions = append(failedRegions, region)
			continue
		}
		entries = append(entries, lo.Map(out.Images, func(image types.Image, _ int) AMIEntry {
			return AMIEntry{
				Name:   lo.FromPtr(image.Name),
				ID:     lo.FromPtr(image.ImageId),
				Region: region,
			}
		})...)
	}

	sort.Slice(entries, func(ii, ji int) bool {
		i := entries[ii]
		j := entries[ji]
		if i.Name != j.Name {
			return i.Name < j.Name
		}
		if i.Region != j.Region {
			return i.Region < j.Region
		}
		return i.ID < j.ID
	})

	if err := w.Write([]string{"ami_name", "ami_id", "region"}); err != nil {
		panic(fmt.Errorf("writing csv, %w", err))
	}
	if err := w.WriteAll(lo.Map(entries, func(entry AMIEntry, _ int) []string {
		return []string{entry.Name, entry.ID, entry.Region}
	})); err != nil {
		panic(fmt.Errorf("writing csv, %w", err))
	}

	fmt.Printf("queried %d AMIs across %d regions\n", len(entries), len(opts.Regions) - len(failedRegions))
	if len(failedRegions) != 0 {
		fmt.Printf("failed to query AMIs for the following regions: %s\n", strings.Join(failedRegions, ", "))
	}
}

func WithRegion(region string) func(*ec2.Options) {
	return func(o *ec2.Options) {
		o.Region = region
	}
}
