# Route53 - Dynamic DNS Updater

So this is kind of neat: _I didn't write most of this go code_; [ChatGPT did](https://chat.openai.com/chat). The absolutely original code [can be found here](https://gist.github.com/danmanners/18de9ca6ed1cf23a1c3ea46b632c2042).

This tool simply checks the Public IP address from where it runs and then updates (or creates) an AWS Route53 A Record. The two flags you'll want to provide are `hostname` and `hosted-zone-id`.

## Usage

You can use the binary like this:

```bash
r53-dyndns \
    --hosted-zone-id Z081932343XEULDET1H8 \
    --hostname your.domain.goes.here
```

## Building and installing the binary

Run these steps on your system

```bash
go mod tidy
go build r53-dyndns.go
sudo chown root:root r53-dyndns
sudo mv r53-dyndns /usr/local/bin/r53-dyndns
```

## Running it as a service

If you'd like to run this tool as a service, you can install the [r53-dyndns.service](r53-dyndns.service) to `/lib/systemd/system/r53-dyndns.service`, make whatever changes you require for your environment, and run:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now r53-dyndns
```

## Recommendation (though _technically_ not a requirement)

To use this tool securely, I recommend creating and IAM user with the permissions of least requirements. The permissions I have my IAM role set to are as such

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "route53:ListTagsForResources",
                "route53:ChangeResourceRecordSets",
                "route53:ListResourceRecordSets",
                "route53:GetHealthCheck",
                "route53:GetHostedZoneLimit",
                "route53:ListTagsForResource"
            ],
            "Resource": "arn:aws:route53:::hostedzone/${HostedZoneIdGoesHere}"
        },
        {
            "Sid": "VisualEditor1",
            "Effect": "Allow",
            "Action": "route53:ListResourceRecordSets",
            "Resource": "arn:aws:route53:::hostedzone/${HostedZoneIdGoesHere}"
        },
        {
            "Sid": "VisualEditor2",
            "Effect": "Allow",
            "Action": [
                "route53:GetAccountLimit",
                "route53:ListTrafficPolicyInstances",
                "route53:GetTrafficPolicyInstanceCount",
                "route53:ListHealthChecks",
                "route53:TestDNSAnswer",
                "route53:ListGeoLocations"
            ],
            "Resource": "*"
        }
    ]
}
```
