import {
  Protobuf,
  main,
} from "https://raw.githubusercontent.com/clementd64/tachiyomi-backup-models/main/mod.ts";

export class TachiqlGenerator extends Protobuf {
  build(): string {
    for (const key in this.defs) {
      for (const entry of this.defs[key]) {
        if (entry.type.startsWith("Backup")) {
          entry.type = entry.type.slice(6);
        }

        if (key === "Backup" && entry.name.startsWith("backup")) {
          if (entry.name === "backupManga") {
            entry.name = "mangas";
          } else {
            entry.name = entry.name[6].toLowerCase() + entry.name.slice(7);
          }
        }
      }

      if (key !== "Backup" && key.startsWith("Backup")) {
        this.defs[key.slice(6)] = this.defs[key];
        delete this.defs[key];
      }
    }

    return super.build();
  }
}

export default TachiqlGenerator;

if (import.meta.main) {
  await main(TachiqlGenerator);
}