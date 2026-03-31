export function canJsonParse(str: any) {
  try {
    const testValueForJson = JSON.parse(str);
    return !!testValueForJson;
  } catch (e) {
    // exception: string is not valid json
    return null;
  }
}

export function getKnownSyntaxHighlighterLanguage(val: any) {
  // to start, only check for json. can expand as needed in future
  const isValidJson = canJsonParse(val);
  return isValidJson ? 'json' : null;
}
