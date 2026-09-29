export const KSA_REGIONS = [
  'Riyadh', 'Makkah', 'Madinah', 'Eastern Province', 'Qassim', 'Asir', 'Tabuk',
  'Hail', 'Northern Borders', 'Jazan', 'Najran', 'Al Bahah', 'Al Jawf',
] as const

export const KSA_CITIES: Record<string, string> = {
  'Riyadh': 'Riyadh', 'Al Kharj': 'Riyadh', 'Diriyah': 'Riyadh', 'Al Majmaah': 'Riyadh', 'Al Zulfi': 'Riyadh',
  'Dawadmi': 'Riyadh', 'Wadi ad-Dawasir': 'Riyadh', 'Afif': 'Riyadh', 'Shaqra': 'Riyadh',
  'Makkah': 'Makkah', 'Jeddah': 'Makkah', 'Taif': 'Makkah', 'Rabigh': 'Makkah', 'Al Qunfudhah': 'Makkah',
  'Al Lith': 'Makkah', 'Khulais': 'Makkah',
  'Madinah': 'Madinah', 'Yanbu': 'Madinah', 'Al Ula': 'Madinah', 'Badr': 'Madinah', 'Khaybar': 'Madinah',
  'Dammam': 'Eastern Province', 'Khobar': 'Eastern Province', 'Dhahran': 'Eastern Province', 'Jubail': 'Eastern Province',
  'Hofuf': 'Eastern Province', 'Qatif': 'Eastern Province', 'Hafr Al-Batin': 'Eastern Province', 'Khafji': 'Eastern Province',
  'Ras Tanura': 'Eastern Province', 'Abqaiq': 'Eastern Province', 'Nairiyah': 'Eastern Province',
  'Buraidah': 'Qassim', 'Unaizah': 'Qassim', 'Ar Rass': 'Qassim', 'Al Bukayriyah': 'Qassim',
  'Abha': 'Asir', 'Khamis Mushait': 'Asir', 'Bisha': 'Asir', 'Al Namas': 'Asir', 'Muhayil': 'Asir',
  'Tabuk': 'Tabuk', 'Duba': 'Tabuk', 'Tayma': 'Tabuk', 'Al Wajh': 'Tabuk', 'Haql': 'Tabuk',
  'Hail': 'Hail', 'Baqaa': 'Hail',
  'Arar': 'Northern Borders', 'Rafha': 'Northern Borders', 'Turaif': 'Northern Borders',
  'Jazan': 'Jazan', 'Sabya': 'Jazan', 'Abu Arish': 'Jazan', 'Samtah': 'Jazan',
  'Najran': 'Najran', 'Sharurah': 'Najran',
  'Al Bahah': 'Al Bahah', 'Baljurashi': 'Al Bahah',
  'Sakaka': 'Al Jawf', 'Dumat al-Jandal': 'Al Jawf', 'Qurayyat': 'Al Jawf',
}

export function regionForCity(city: string): string | undefined {
  const wanted = city.trim().toLowerCase()
  const match = Object.keys(KSA_CITIES).find((name) => name.toLowerCase() === wanted)
  return match ? KSA_CITIES[match] : undefined
}
