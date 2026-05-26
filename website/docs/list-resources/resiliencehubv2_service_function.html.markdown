---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_service_function"
description: |-
  Lists Resilience Hub V2 Service Function resources.
---

# List Resource: aws_resiliencehubv2_service_function

Lists Resilience Hub V2 Service Function resources.

## Example Usage

```terraform
list "aws_resiliencehubv2_service_function" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
