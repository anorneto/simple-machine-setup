import logging
from pathlib import Path
import subprocess
from typing import Any, Final
from src.enums import SystemType
from logging import getLogger
import tomllib

logger = getLogger(__name__)

# configure logger to print to console
logger.setLevel("INFO")
console_handler = logging.StreamHandler()
console_handler.setLevel("INFO")
formatter = logging.Formatter("%(asctime)s - %(levelname)s - %(message)s")
console_handler.setFormatter(formatter)
logger.addHandler(console_handler)


class SystemSetupService:
    CONFIGS_FOLDER: Final[Path] = Path("configs")

    def __init__(self) -> None:
        system_type: SystemType = SystemType.from_platform()
        logger.info("Initializing Setup for %s", system_type.value.upper())
        config_data: dict[str, Any] = self._parse_system_configs(
            configs_folder=self.CONFIGS_FOLDER, system_type=system_type
        )

        self._install_from_config(config_data=config_data)

    def _parse_system_configs(self, configs_folder: Path, system_type: SystemType) -> dict[str, Any]:
        config_file: Path = configs_folder / f"{system_type.value}_config.toml"
        if not config_file.exists():
            raise FileNotFoundError(f"Configuration file {config_file} not found.")

        logger.info("Parsing configuration from %s", config_file)
        # read toml file and parse it
        with config_file.open("rb") as file:
            try:
                config_data: dict[str, Any] = tomllib.load(file)
                logger.info("Configuration loaded successfully.")
            except tomllib.TOMLDecodeError as e:
                logger.error("Failed to parse configuration file: %s", e)
                raise

        return config_data

    def _install_from_config(self, config_data: dict[str, Any]) -> None:
        logger.info("Installing packages from configuration data.")

        for install in config_data.get("custom_install", {}).get("commands", []):
            name = install.get("name")
            command = install.get("command")

            result = subprocess.run(["which", name], capture_output=True)
            if result.returncode == 0:
                logger.info("%s already installed, skipping.", name)
                # continue
            logger.info("> Installing package: %s", name)
            subprocess.run(command, shell=True, check=True)
