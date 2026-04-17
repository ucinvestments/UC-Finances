export interface CampusTheme {
  primary: string;
  secondary: string;
  accent: string;
  light: string;
  dark: string;
  gradient: string;
  name: string;
}

export const campusThemes: Record<string, CampusTheme> = {
  Berkeley: {
    primary: "#003262",
    secondary: "#FDB515",
    accent: "#859438",
    light: "#C4820E",
    dark: "#001B44",
    gradient: "linear-gradient(135deg, #003262 0%, #FDB515 100%)",
    name: "UC Berkeley",
  },
  "Los Angeles": {
    primary: "#2774AE",
    secondary: "#FFD100",
    accent: "#8BB8E8",
    light: "#FFED4E",
    dark: "#003B5C",
    gradient: "linear-gradient(135deg, #2774AE 0%, #FFD100 100%)",
    name: "UCLA",
  },
  "San Diego": {
    primary: "#182B49",
    secondary: "#C69214",
    accent: "#6F99D3",
    light: "#E8E3D0",
    dark: "#0F1419",
    gradient: "linear-gradient(135deg, #182B49 0%, #C69214 100%)",
    name: "UC San Diego",
  },
  Davis: {
    primary: "#022851",
    secondary: "#FFBF00",
    accent: "#B0D236",
    light: "#C6AA76",
    dark: "#001633",
    gradient: "linear-gradient(135deg, #022851 0%, #FFBF00 100%)",
    name: "UC Davis",
  },
  Irvine: {
    primary: "#0064A4",
    secondary: "#FFD200",
    accent: "#1B365D",
    light: "#7FC8F8",
    dark: "#003D5C",
    gradient: "linear-gradient(135deg, #0064A4 0%, #FFD200 100%)",
    name: "UC Irvine",
  },
  "Santa Barbara": {
    primary: "#003660",
    secondary: "#DDD15A",
    accent: "#9CBEBE",
    light: "#F2E6CE",
    dark: "#001B2E",
    gradient: "linear-gradient(135deg, #003660 0%, #DDD15A 100%)",
    name: "UC Santa Barbara",
  },
  "Santa Cruz": {
    primary: "#003C6C",
    secondary: "#F29813",
    accent: "#7BA7BC",
    light: "#FFEAA7",
    dark: "#002147",
    gradient: "linear-gradient(135deg, #003C6C 0%, #F29813 100%)",
    name: "UC Santa Cruz",
  },
  Riverside: {
    primary: "#003DA5",
    secondary: "#FFB81C",
    accent: "#005581",
    light: "#B7CDF1",
    dark: "#002952",
    gradient: "linear-gradient(135deg, #003DA5 0%, #FFB81C 100%)",
    name: "UC Riverside",
  },
  "San Francisco": {
    primary: "#052049",
    secondary: "#90BD31",
    accent: "#18A3AC",
    light: "#6DACE4",
    dark: "#031329",
    gradient: "linear-gradient(135deg, #052049 0%, #90BD31 100%)",
    name: "UCSF",
  },
  Merced: {
    primary: "#002856",
    secondary: "#FFB310",
    accent: "#8A8B8C",
    light: "#B3A369",
    dark: "#001B3D",
    gradient: "linear-gradient(135deg, #002856 0%, #FFB310 100%)",
    name: "UC Merced",
  },
  UCOP: {
    primary: "#1B4685",
    secondary: "#DAAA00",
    accent: "#767F8B",
    light: "#8BB8E8",
    dark: "#0D2142",
    gradient: "linear-gradient(135deg, #1B4685 0%, #DAAA00 100%)",
    name: "UC Office of the President",
  },
};

const SNAKE_TO_DISPLAY: Record<string, string> = {
  berkeley: "Berkeley",
  davis: "Davis",
  irvine: "Irvine",
  los_angeles: "Los Angeles",
  merced: "Merced",
  riverside: "Riverside",
  san_diego: "San Diego",
  san_francisco: "San Francisco",
  santa_barbara: "Santa Barbara",
  santa_cruz: "Santa Cruz",
  ucop: "UCOP",
};

const DEFAULT_THEME: CampusTheme = {
  primary: "#1B4685",
  secondary: "#DAAA00",
  accent: "#767F8B",
  light: "#8BB8E8",
  dark: "#0D2142",
  gradient: "linear-gradient(135deg, #1B4685 0%, #DAAA00 100%)",
  name: "Unknown",
};

export function getCampusTheme(campus: string): CampusTheme {
  const display = SNAKE_TO_DISPLAY[campus] ?? campus;
  return campusThemes[display] ?? { ...DEFAULT_THEME, name: campus };
}

export function campusDisplayName(snakeKey: string): string {
  return SNAKE_TO_DISPLAY[snakeKey] ?? snakeKey;
}

export function getCampusColorScale() {
  const colors = Object.values(campusThemes).map((t) => t.primary);
  const campuses = Object.keys(campusThemes);
  return { colors, campuses };
}
