package cloud.luxor.lbwl.flash

import org.bukkit.Location

data class Checkpoint(val location: Location, val timeReached: Long)

data class GameData(val checkpoints: MutableSet<Checkpoint>, val mapSpeed: Int, val spawn: Location)
