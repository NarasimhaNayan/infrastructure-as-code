#!/bin/bash
# security-scan.sh
# Usage: ./security-scan.sh <image-name>
# This script scans the specified Docker image using Trivy and exits with a non-zero code if vulnerabilities of severity HIGH or CRITICAL are found.

IMAGE=$1

echo "Starting security scan for image: $IMAGE"

# Run Trivy scan
trivy image --exit-code 1 --severity HIGH,CRITICAL $IMAGE
SCAN_RESULT=$?

if [ $SCAN_RESULT -ne 0 ]; then
    echo "Security scan found vulnerabilities with HIGH or CRITICAL severity."
    exit $SCAN_RESULT
else
    echo "Security scan passed with no HIGH or CRITICAL vulnerabilities."
    exit 0
fi
