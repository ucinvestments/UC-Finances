# libraries necessary: xmltodict, secedgar, pandas

from secedgar import filings, FilingType

filing_types = [FilingType.FILING_13FHR_AMEND,
                                  FilingType.FILING_13FHR,
                                  FilingType.FILING_SC13G,
                                  FilingType.FILING_SC13G_AMEND]

filing_names = ["13F-HR_A", "13F-HR", "SC_13G", "SC_13G_A"]

for i in range(4):
  my_filings = filings(cik_lookup='0000315054',
                      user_agent="admin@ucinvestments.info",
                      filing_type=filing_types[i])

  my_filings_urls = my_filings.get_urls()
  my_filings.save(f"/{filing_names[i]}")
