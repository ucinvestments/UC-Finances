import os
import re
import xmltodict
import pandas as pd

folder = "/0000315054/13F-HR/"






def extract_table(start_word, end_word, text, group):

  # find where table is stored in file (using keywords) + figure out what format it's stored in
  pattern = f"({start_word})(.*?)({end_word})"
  matches = re.findall(pattern, text, re.DOTALL)

  # parse XML format
  if start_word == "<XML>" and matches:
    matches_no_intro = matches[1]
    match_dict = xmltodict.parse(matches_no_intro[1].strip())
    info_table = match_dict["informationTable"]["infoTable"]
    df = pd.DataFrame.from_dict(info_table)

    df['sshPrnamt'] = df['shrsOrPrnAmt'].apply(lambda x: x['sshPrnamt'])
    df['sshPrnamtType'] = df['shrsOrPrnAmt'].apply(lambda x: x['sshPrnamtType'])


    df['Sole'] = df['votingAuthority'].apply(lambda x: x['Sole'])
    df['Shared'] = df['votingAuthority'].apply(lambda x: x['Shared'])
    df['None'] = df['votingAuthority'].apply(lambda x: x['None'])

    df = df.drop(columns=['shrsOrPrnAmt', 'votingAuthority'])

    return df

  # parse TSV-like format
  elif start_word == "<TABLE>" and matches:
    for match in matches:
      # the formatting of files before 2013 is so inconsistent from year to year that
      # i don't know if i'm going to put the effort in to do it.
      # if you do
      # then you're awesome
      return match[1]





# parse through files in folder
for file in os.listdir(folder):
  file_location = os.path.join(folder, file)
  f = open(file_location, 'r')
  file_content = f.read()

  if ("<XML>" in file_content):
    print(f"{file} contains an XML file")
    start_word = "<XML>"
    end_word = "</XML>"
    xml_csv = extract_table(start_word, end_word, file_content, 1)
    xml_csv.to_csv(folder + f'{file[:-4]}.csv', index=False)

  elif ("<TABLE>" in file_content):
    print(f"{file} contains an tsv file")
    start_word = "<TABLE>"
    end_word = "</TABLE>"
    tsv_table = extract_table(start_word, end_word, file_content, 2)
    print(tsv_table)

  else:
    print(f"hmm idk what this is: {file}")
  f.close()
