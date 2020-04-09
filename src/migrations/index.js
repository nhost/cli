const path = require("path");
const fs = require("fs");

module.exports = function (destination) {
  let migrationsFolder = path.resolve(
    path.dirname(require.main.filename),
    "../src/migrations"
  );

  const now = Date.now();
  fs.mkdirSync(`${destination}/${now}_init`, { recursive: true });

  fs.copyFileSync(`${migrationsFolder}/up.sql`, `${destination}/${now}_init/up.sql`);
  fs.copyFileSync(`${migrationsFolder}/up.yaml`, `${destination}/${now}_init/up.yaml`);
};
