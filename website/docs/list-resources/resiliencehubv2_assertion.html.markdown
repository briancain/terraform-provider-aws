---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_assertion"
description: |-
  Lists Resilience Hub V2 Assertion resources.
---

# List Resource: aws_resiliencehubv2_assertion

Lists Resilience Hub V2 Assertion resources.

## Example Usage

```terraform
list "aws_resiliencehubv2_assertion" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
