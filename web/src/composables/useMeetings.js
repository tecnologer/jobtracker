import { ref } from 'vue'
import * as api from '../api'

const upcomingMeetings = ref([])

export function useMeetings() {
  async function loadUpcomingMeetings() {
    upcomingMeetings.value = await api.fetchUpcomingMeetings()
  }

  return { upcomingMeetings, loadUpcomingMeetings }
}
