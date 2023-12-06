package cloud.luxor.lbwl.flash

import com.charleskorn.kaml.PolymorphismStyle
import com.charleskorn.kaml.Yaml
import kotlinx.serialization.Serializable
import kotlinx.serialization.decodeFromString
import org.bukkit.Location
import org.bukkit.Material
import org.bukkit.World
import org.slf4j.Logger
import java.io.File


@Serializable
data class MapConfig(
    val name: String = "null",
    val checkpoints: Int,
    val author: String = "",
    val time: Int? = null,
    val mode: String = "easy",
    val speedLevel: Int = 19,
    val item: Material = Material.STONE,
    val spawn: String = "0,0,0,0"
) {
    companion object {

        /**
         * read a mapconfig.yml file for a flash map
         *
         * @param file the mapconfig.yml file
         * @param logger optional for logging the current name of the map
         * @return a result of the mapconfig, one should explicitly handle the exceptions that happens during parsing
         */
        fun read(file: File, logger: Logger? = null): Result<MapConfig> {
            return try {
                val conf = createYamlParser().decodeFromString<MapConfig>(file.readText())
                logger?.info("Loaded map '${conf.name}'")
                Result.success(conf)
            } catch (e: Exception) {
                logger?.warn("Failed to load map config: ${file.name}")
                Result.failure(e)
            }
        }

        // stupid legacy location format
        fun locFromString(rawLocation: String, world: World): Location? {
            val pos = rawLocation
                .split(",")
                .map { it.toDouble() }
            if (pos.size < 3)
                return null
            val location = Location(world, pos[0], pos[1], pos[2])
            location.yaw = pos
                .getOrElse(3) { 0 }
                .toFloat() * 90
            return location
        }
    }
}


private fun createYamlParser(): Yaml {
    val yamlConfig = Yaml.default.configuration.copy(
        polymorphismStyle = PolymorphismStyle.Property,
        polymorphismPropertyName = "type",
    )
    return Yaml(Yaml.default.serializersModule, yamlConfig)
}
