package cloud.luxor.lbwl.flash

import net.kyori.adventure.text.Component
import net.kyori.adventure.text.TextComponent
import net.kyori.adventure.text.format.Style
import net.kyori.adventure.text.format.TextColor
import org.bukkit.Bukkit
import org.bukkit.entity.Player
import org.bukkit.event.EventHandler
import org.bukkit.event.Listener
import org.bukkit.event.inventory.InventoryClickEvent
import org.bukkit.event.player.PlayerQuitEvent
import org.bukkit.inventory.Inventory
import org.bukkit.inventory.ItemStack
import org.bukkit.plugin.Plugin
import java.awt.Color
import java.io.File

/**
 * @author Yannic Rieger
 */
@Suppress("unused")
class MapVoting(private val maps: List<Pair<MapConfig, File>>) : Listener {

    private val votes = HashMap<String, Int>()
    private val playerVotes = HashMap<Player, String>()
    private val inventory: Inventory = Bukkit.createInventory(null, 5 * 9, Component.text("Maps"))

    private fun updateInventory() {
        // Probably pretty inefficient for sorting by difficulty, but it does the job and 
        // efficiency doesn't really matter in this context.
        // Maps with easy difficulty should be listed first ('e' before 'h') 
        val sortedMaps = this.maps.sortedWith(compareBy { it.first.mode.first() })
        sortedMaps.forEachIndexed { index, (mapConfig, _) ->
            val stack = ItemStack(mapConfig.item, votes[mapConfig.name] ?: 1)   //0 is forbidden in 1.19
            val meta = stack.itemMeta
            val displayColor = if (mapConfig.mode == "easy") Color.GREEN.rgb else Color.RED.rgb
            meta.displayName(
                Component
                    .text(mapConfig.name.replace("-", " "))
                    .style(Style.style(TextColor.color(displayColor)))
            )
            meta.lore(
                listOf(
                    Component
                        .text("Erbauer: §b${mapConfig.author}")
                        .style(Style.style(TextColor.color(Color.YELLOW.rgb)))
                )
            )
            stack.itemMeta = meta
            inventory.setItem(index, stack)
        }
    }

    init {
        this.updateInventory()
    }

    private fun vote(player: Player, name: String) {
        if (this.playerVotes.containsKey(player)) {
            this.removeVote(player)
        }

        this.playerVotes[player] = name

        if (!this.votes.containsKey(name)) this.votes[name] = 1

        this.votes[name] = this.votes[name]!!.plus(1)

        this.updateInventory()
    }

    private fun removeVote(player: Player) {
        if (!this.playerVotes.containsKey(player)) return
        val map = this.playerVotes[player]!!
        this.votes[map] = this.votes[map]!!.minus(1)
        this.updateInventory()
    }

    fun open(player: Player) {
        player.openInventory(this.inventory)
    }

    fun determineMap(): Pair<MapConfig, File> {
        // choose random map if no-one has voted
        if (this.votes.values.sum() == 0) return this.maps.shuffled()[0]

        // find the map with the highest votes
        val name = this.votes.toList().maxBy { (_, value) -> value }
        return this.maps.find { (config) -> name.first.contains(config.name) }!!
    }

    @EventHandler
    private fun onInventoryClick(event: InventoryClickEvent) {
        if (event.clickedInventory == null) return
        //if((event.clickedInventory?.holder as Container).name != "Text")
        if (event.currentItem == null) return
        if (event.currentItem!!.itemMeta == null) return

        event.isCancelled = true

        val name = event.currentItem?.itemMeta?.displayName() as TextComponent
        name.let { this.vote(event.whoClicked as Player, it.content()) }
    }

    @EventHandler
    private fun onPlayerQuit(event: PlayerQuitEvent) {
        this.removeVote(event.player)
    }

    fun registerListeners(plugin: Plugin) {
        Bukkit.getPluginManager().registerEvents(this, plugin)
    }

    fun end() {
        this.votes.clear()
        this.playerVotes.clear()
        InventoryClickEvent.getHandlerList().unregister(this)
    }
}
