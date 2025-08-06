from enum import StrEnum

import platform


class SystemType(StrEnum):
    """Enumeration for system types."""

    LINUX = "linux"
    WINDOWS = "windows"
    MACOS = "macos"

    @classmethod
    def from_platform(cls) -> "SystemType":
        """Get the system type based on the current platform."""
        system = platform.system().lower()
        if system == "linux":
            return cls.LINUX
        elif system == "windows":
            return cls.WINDOWS
        elif system == "darwin":
            return cls.MACOS
        else:
            raise ValueError(f"Unsupported system type: {system}")
