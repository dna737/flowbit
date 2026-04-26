/**
 * Anonymous tab-scoped id (`word-word-word`), sessionStorage only.
 */

const STORAGE_KEY = "flowbit:user_id";

const ADJ = [
  "calm",
  "quick",
  "lazy",
  "bold",
  "wise",
  "cool",
  "wild",
  "quiet",
  "brave",
  "fancy",
  "lucky",
  "noble",
  "proud",
  "salty",
  "rusty",
  "jolly",
  "mighty",
  "nimble",
  "royal",
  "stale",
  "vivid",
  "zesty",
  "humble",
  "gentle",
  "random",
  "curly",
  "fuzzy",
  "happy",
  "sleepy",
  "grumpy",
] as const;

const NOUN_A = [
  "horse",
  "river",
  "cloud",
  "tiger",
  "paper",
  "clock",
  "bridge",
  "castle",
  "dragon",
  "eagle",
  "forest",
  "garden",
  "hammer",
  "island",
  "jungle",
  "kitten",
  "meadow",
  "nebula",
  "ocean",
  "penguin",
  "quartz",
  "rocket",
  "sphinx",
  "tunnel",
  "valley",
  "wizard",
  "pixel",
  "ember",
  "falcon",
  "comet",
] as const;

const NOUN_B = [
  "radish",
  "pickle",
  "walnut",
  "muffin",
  "pretzel",
  "biscuit",
  "caramel",
  "noodle",
  "turnip",
  "parsley",
  "wasabi",
  "kimchi",
  "sprocket",
  "widget",
  "gadget",
  "gizmo",
  "socket",
  "bucket",
  "harvest",
  "harbor",
  "pebble",
  "marble",
  "granite",
  "opal",
  "ember",
  "falcon",
  "meadow",
  "canyon",
  "comet",
  "meteor",
] as const;

function pick<T extends readonly string[]>(arr: T): T[number] {
  return arr[Math.floor(Math.random() * arr.length)]!;
}

function mint(): string {
  return `${pick(ADJ)}-${pick(NOUN_A)}-${pick(NOUN_B)}`;
}

/** Tab-scoped anonymous id for `X-User-Id`. Override with `VITE_USER_ID` for tests. */
export function getUserId(): string {
  const forced = import.meta.env.VITE_USER_ID;
  if (typeof forced === "string" && forced.trim() !== "") {
    return forced.trim();
  }
  try {
    let id = sessionStorage.getItem(STORAGE_KEY);
    if (!id?.trim()) {
      id = mint();
      sessionStorage.setItem(STORAGE_KEY, id);
    }
    return id.trim();
  } catch {
    return mint();
  }
}
