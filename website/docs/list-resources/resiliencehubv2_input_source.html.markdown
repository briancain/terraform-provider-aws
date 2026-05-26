---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_input_source"
description: |-
  Lists Resilience Hub V2 Input Source resources.
---

# List Resource: aws_resiliencehubv2_input_source

Lists Resilience Hub V2 Input Source resources.

## Example Usage

```terraform
list "aws_resiliencehubv2_input_source" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
