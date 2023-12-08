package cloud.luxor.lbwl.flash.listener

import cloud.luxor.lbwl.flash.*
import cloud.luxor.lbwl.flash.event.PlayerCheckpointEvent
import cloud.luxor.lbwl.flash.event.PlayerFinishedEvent
import net.kyori.adventure.text.Component
import org.bukkit.Bukkit
import org.bukkit.GameMode
import org.bukkit.Material
import org.bukkit.Sound
import org.bukkit.block.Furnace
import org.bukkit.block.Sign
import org.bukkit.entity.Player
import org.bukkit.event.EventHandler
import org.bukkit.event.Listener
import org.bukkit.event.block.Action
import org.bukkit.event.entity.EntityDamageEvent
import org.bukkit.event.entity.PlayerDeathEvent
import org.bukkit.event.player.PlayerInteractEvent
import org.bukkit.event.player.PlayerJoinEvent
import org.bukkit.event.player.PlayerLoginEvent
import org.bukkit.inventory.InventoryView

@Suppress("unused")
class PlayerListener(private val plugin: FlashPlugin) : Listener {

    @EventHandler
    private fun onFinishTriggered(event: PlayerInteractEvent) {
        if (event.player.gameMode == GameMode.SPECTATOR) return
        if (event.action != Action.PHYSICAL) return
        if (event.clickedBlock?.type != Material.OAK_PRESSURE_PLATE) return
        val type = event.clickedBlock?.location?.subtract(0.0, 1.0, 0.0)?.block?.type
        if (!type.toString().contains("WOOL")) return
        Bukkit.getPluginManager().callEvent(PlayerFinishedEvent(event.player, System.currentTimeMillis()))
    }

    @EventHandler
    private fun onCheckpointTriggered(event: PlayerInteractEvent) {
        if (event.player.gameMode == GameMode.SPECTATOR) {
            event.isCancelled = true
            return
        }
        if (event.action != Action.PHYSICAL) return
        if (event.clickedBlock?.type != Material.STONE_PRESSURE_PLATE) return

        // Legacy checkpoint format
        val player = event.player
        val state = event.clickedBlock!!.location.subtract(0.0, 1.0, 0.0).block.state

        if (state !is Furnace) return

        val number = state.inventory.contents.sumOf {
            it?.amount ?: 0
        }

        if (number == player.getCurrentCheckPointIndex() + 1) {
            Bukkit.getPluginManager().callEvent(
                    PlayerCheckpointEvent(player, Checkpoint(player.location.clone(), System.currentTimeMillis()))
            )
            return
        }

        if (number > player.getCurrentCheckPointIndex() + 1) {
            player.sendMessage("$PREFIX §cDu hast ein Checkpoint übersprungen! Du wurdest zurück teleportiert!")
            player.respawn()
            player.playSound(player.location, Sound.ENTITY_ENDERMAN_DEATH, 1.0F, 1.0F)
            return
        }
    }

    @EventHandler
    private fun onPlayerJoin(event: PlayerJoinEvent) {
        val player = event.player
        player.inventory.clear()
        if (this.plugin.state != GameState.WAITING) {
            player.activePotionEffects.forEach { player.removePotionEffect(it.type) }
            player.gameMode = GameMode.SPECTATOR
            this.plugin.mapSpawnLocation?.let { player.teleport(it) }
            return
        }
        player.gameMode = GameMode.ADVENTURE
        player.applyEffects()
        player.teleport(this.plugin.spawnLocation)
        player.inventory.setItem(4, create(Material.COMPASS, 0, "§a§lMap-Auswahl"))
    }

    @EventHandler
    private fun onPlayerLogin(event: PlayerLoginEvent) {
        if (this.plugin.state != GameState.INIT) return
        event.disallow(PlayerLoginEvent.Result.KICK_OTHER, Component.text("§cServer initialisiert..."))
    }

    @EventHandler
    private fun onItemInteract(event: PlayerInteractEvent) {
        if (event.action != Action.RIGHT_CLICK_AIR && event.action != Action.RIGHT_CLICK_BLOCK) return
        if (event.item == null) return

        val type = event.item!!.type
        val player = event.player

        if (type == Material.RED_DYE) {
            player.respawn()
            return
        }

        if (type == Material.BLAZE_ROD || type == Material.STICK) {
            event.player.toggleVisibility(plugin)
            return
        }

        if (type == Material.COMPASS) {
            this.plugin.mapVoting.open(player)
            return
        }
    }

    @EventHandler
    private fun onPlayerDeath(event: PlayerDeathEvent) {
        event.keepInventory = true
        event.entity.player?.spigot()?.respawn()
        Bukkit.getScheduler().runTaskLater(plugin, Runnable { event.entity.respawn() }, 1)
        event.deathMessage(Component.text(""))
    }

    @EventHandler
    private fun onEntityDamage(event: EntityDamageEvent) {
        if (event.entity !is Player) return
        val player = event.entity as Player

        if (!player.isIngame()) event.isCancelled = true
        if (event.cause != EntityDamageEvent.DamageCause.VOID) return

        player.respawn()
    }
}
