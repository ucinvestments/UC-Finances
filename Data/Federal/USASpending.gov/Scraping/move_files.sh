#!/usr/bin/env bash

# Script to move scraped award files to their respective directories
# Run this after the scraper completes

echo "Moving award files to respective directories..."

# Get the date from the filenames (assumes today's date)
DATE=$(date +%Y-%m-%d)

# Move files with proper permissions
if [ -f "uc_contracts_${DATE}.json" ]; then
    echo "Moving contracts..."
    sudo mv "uc_contracts_${DATE}.json" "../Contracts/"
    sudo chown okita:users "../Contracts/uc_contracts_${DATE}.json"
fi

if [ -f "uc_grants_${DATE}.json" ]; then
    echo "Moving grants..."
    sudo mv "uc_grants_${DATE}.json" "../Grants/"
    sudo chown okita:users "../Grants/uc_grants_${DATE}.json"
fi

if [ -f "uc_loans_${DATE}.json" ]; then
    echo "Moving loans..."
    sudo mv "uc_loans_${DATE}.json" "../Loans/"
    sudo chown okita:users "../Loans/uc_loans_${DATE}.json"
fi

if [ -f "uc_idvs_${DATE}.json" ]; then
    echo "Moving IDVs..."
    sudo mv "uc_idvs_${DATE}.json" "../Contract_IDVs/"
    sudo chown okita:users "../Contract_IDVs/uc_idvs_${DATE}.json"
fi

if [ -f "uc_other_financial_assistance_${DATE}.json" ]; then
    echo "Moving other financial assistance..."
    sudo mv "uc_other_financial_assistance_${DATE}.json" "../Other_Financial_Assistance/"
    sudo chown okita:users "../Other_Financial_Assistance/uc_other_financial_assistance_${DATE}.json"
fi

if [ -f "uc_direct_payments_${DATE}.json" ]; then
    echo "Moving direct payments..."
    sudo mv "uc_direct_payments_${DATE}.json" "../Direct_Payments/"
    sudo chown okita:users "../Direct_Payments/uc_direct_payments_${DATE}.json"
fi

echo "File organization complete!"
echo "Files moved to their respective award type directories."
