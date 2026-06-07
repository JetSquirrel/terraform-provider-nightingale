# Complete Example

This example demonstrates a complete Nightingale monitoring setup using Terraform.

It creates:
- A notification rule for email alerts
- Two alert rules (disk critical and CPU warning)
- An alert subscription that routes critical alerts to the ops team

## Usage

```shell
terraform init
terraform plan -var="nightingale_endpoint=https://n9e.example.com" -var="nightingale_token=your-token"
terraform apply -var="nightingale_endpoint=https://n9e.example.com" -var="nightingale_token=your-token"
```
