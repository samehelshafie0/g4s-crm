// KSA Regions and Cities Data

export interface KSARegion {
  name: string
  cities: string[]
}

export const KSA_REGIONS: KSARegion[] = [
  {
    name: 'Riyadh',
    cities: [
      'Riyadh',
      'Diriyah',
      'Al-Kharj',
      'Dawadmi',
      'Al-Majmaah',
      'Al-Quwayiyah',
      'Afif',
      'Al-Zulfi',
      'Shaqra',
      'Hotat Bani Tamim',
      'Al-Hariq',
      'Al-Ghat',
      'Rumah'
    ]
  },
  {
    name: 'Makkah',
    cities: [
      'Makkah',
      'Jeddah',
      'Taif',
      'Rabigh',
      'Khulais',
      'Rania',
      'Turubah',
      'Jumum',
      'Al-Kamil',
      'Al-Muwayh',
      'Maysan',
      'Adham',
      'Al-Lith',
      'Al-Qunfudhah'
    ]
  },
  {
    name: 'Madinah',
    cities: [
      'Madinah',
      'Yanbu',
      'Al-Ula',
      'Badr',
      'Khaybar',
      'Al-Mahd',
      'Al-Hanakiyah',
      'Wadi Al-Fara'
    ]
  },
  {
    name: 'Eastern Province',
    cities: [
      'Dammam',
      'Dhahran',
      'Khobar',
      'Jubail',
      'Al-Ahsa',
      'Hofuf',
      'Mubarraz',
      'Qatif',
      'Safwa',
      'Ras Tanura',
      'Khafji',
      'Abqaiq',
      'Nairiyah',
      'Al-Qaisumah'
    ]
  },
  {
    name: 'Asir',
    cities: [
      'Abha',
      'Khamis Mushait',
      'Bisha',
      'Muhayil',
      'Bareq',
      'Sarat Abidah',
      'Ahad Rafidah',
      'Al-Namas',
      'Tathlith',
      'Dhahran Al-Janoub'
    ]
  },
  {
    name: 'Tabuk',
    cities: [
      'Tabuk',
      'Al-Wajh',
      'Duba',
      'Tayma',
      'Haql',
      'Al-Bad',
      'Umluj'
    ]
  },
  {
    name: 'Qassim',
    cities: [
      'Buraidah',
      'Unaizah',
      'Al-Rass',
      'Al-Mithnab',
      'Al-Badaya',
      'Al-Bukayriyah',
      'Riyadh Al-Khabra',
      'Al-Asyah',
      'Al-Nabhaniyah',
      'Uqlat Al-Suqur',
      'Al-Shammasiyah',
      'Dhurma'
    ]
  },
  {
    name: 'Hail',
    cities: [
      'Hail',
      'Baqaa',
      'Al-Ghazalah',
      'Al-Shamli',
      'Jubbah',
      'Al-Hayt',
      'Al-Sulaymi',
      'Samira',
      'Mawqaq'
    ]
  },
  {
    name: 'Northern Borders',
    cities: [
      'Arar',
      'Rafha',
      'Turaif',
      'Al-Uwayqilah'
    ]
  },
  {
    name: 'Jazan',
    cities: [
      'Jazan',
      'Sabya',
      'Abu Arish',
      'Samtah',
      'Farasan',
      'Damad',
      'Baish',
      'Al-Darb',
      'Al-Aydabi',
      'Al-Rith',
      'Ahad Al-Masarihah',
      'Al-Tuwal'
    ]
  },
  {
    name: 'Najran',
    cities: [
      'Najran',
      'Sharurah',
      'Hubuna',
      'Badr Al-Janoub',
      'Yadamah',
      'Kharkhir',
      'Thar'
    ]
  },
  {
    name: 'Al-Bahah',
    cities: [
      'Al-Bahah',
      'Baljurashi',
      'Al-Mandaq',
      'Al-Mikhwah',
      'Qilwah',
      'Al-Aqiq',
      'Al-Qara',
      'Bani Hassan',
      'Ghamid Al-Zinad'
    ]
  },
  {
    name: 'Al-Jouf',
    cities: [
      'Sakakah',
      'Qurayyat',
      'Dumat Al-Jandal',
      'Tabarjal'
    ]
  }
]

// Get all cities from all regions
export const getAllCities = (): string[] => {
  const allCities: string[] = []
  KSA_REGIONS.forEach(region => {
    allCities.push(...region.cities)
  })
  return allCities.sort()
}

// Get cities for specific regions
export const getCitiesForRegions = (regions: string[]): string[] => {
  const cities: string[] = []
  regions.forEach(regionName => {
    const region = KSA_REGIONS.find(r => r.name === regionName)
    if (region) {
      cities.push(...region.cities)
    }
  })
  return cities.sort()
}

// Get all region names
export const getAllRegionNames = (): string[] => {
  return KSA_REGIONS.map(r => r.name)
}
