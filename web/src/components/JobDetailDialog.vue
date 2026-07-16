<template>
  <BaseDialog
    sheet
    width="w-full h-full overflow-y-auto md:w-[min(90vw,42rem)] md:h-auto md:max-h-[90vh] lg:w-[62vw]"
    @close="emit('close')"
  >
    <div class="flex justify-between items-start mb-4 gap-3">
      <div class="flex-1 flex flex-col gap-2">
        <div class="flex flex-col md:flex-row gap-2">
          <div class="flex flex-col gap-1 flex-1">
            <label class="text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('common.company') }}</label>
            <input
              v-model="edit.company"
              :placeholder="t('common.company')"
              class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-1.5 text-sm font-semibold bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
          </div>
          <div class="flex flex-col gap-1 flex-1">
            <label class="text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('common.position') }}</label>
            <input
              v-model="edit.position"
              :placeholder="t('common.position')"
              class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-1.5 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
          </div>
        </div>
        <div class="flex flex-col gap-2 md:flex-row md:items-end">
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('common.status') }}</label>
            <select
              v-model="edit.status"
              class="border border-gray-300 dark:border-gray-600 rounded-lg px-2 py-1 text-xs bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option
                v-for="s in statuses"
                :key="s"
                :value="s"
              >
                {{ t('status.' + s) }}
              </option>
            </select>
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('common.stage') }}</label>
            <div class="flex items-center gap-1">
              <select
                v-model="edit.stage_id"
                class="border border-gray-300 dark:border-gray-600 rounded-lg px-2 py-1 text-xs bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option :value="0">
                  {{ t('common.noStage') }}
                </option>
                <option
                  v-for="s in stages"
                  :key="s.id"
                  :value="s.id"
                >
                  {{ tStage(s.name) }}
                </option>
              </select>
              <button
                :title="t('detail.manageStagesTooltip')"
                class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors leading-none"
                @click="stagesMgmt = true"
              >
                &#9881;
              </button>
            </div>
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('common.applied') }}</label>
            <input
              v-model="edit.applied_at"
              type="date"
              class="border border-gray-300 dark:border-gray-600 rounded-lg px-2 py-1 text-xs bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
          </div>
        </div>
      </div>
      <div class="flex items-center gap-2 shrink-0">
        <button
          :title="job.top_match ? t('jobs.removeTopMatch') : t('jobs.markTopMatch')"
          :aria-label="job.top_match ? t('jobs.removeTopMatch') : t('jobs.markTopMatch')"
          class="min-h-11 min-w-11 md:min-h-0 md:min-w-0 inline-flex items-center justify-center"
          @click="toggleTopMatch(job)"
        >
          <svg
            :class="job.top_match ? 'text-amber-500 fill-current' : 'text-gray-400 dark:text-gray-500 fill-none stroke-current'"
            class="w-5 h-5"
            viewBox="0 0 20 20"
            stroke-width="1.5"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
            />
          </svg>
        </button>
        <button
          class="min-h-11 min-w-11 md:min-h-0 md:min-w-0 inline-flex items-center justify-center text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-lg leading-none"
          @click="emit('close')"
        >
          ✕
        </button>
      </div>
    </div>

    <div class="mb-4 pb-4 border-b border-gray-100 dark:border-gray-700 flex flex-col gap-2">
      <div class="flex flex-col gap-1">
        <label class="text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('common.notes') }}</label>
        <textarea
          v-model="edit.notes"
          rows="2"
          :placeholder="t('common.notes')"
          class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
        />
      </div>
      <div class="flex flex-col gap-1">
        <label class="text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('detail.jobUrl') }}</label>
        <div class="flex gap-2 items-center">
          <input
            v-model="edit.url"
            type="url"
            :placeholder="t('detail.urlPlaceholder')"
            class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
          <a
            :href="edit.url || '#'"
            target="_blank"
            rel="noopener"
            :class="edit.url ? 'text-blue-500 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 cursor-pointer' : 'text-blue-500 dark:text-blue-400 opacity-25 pointer-events-none'"
            :aria-disabled="!edit.url"
            :title="t('detail.openJobUrl')"
            class="text-lg leading-none transition-colors shrink-0"
          >&#128279;</a>
        </div>
      </div>
    </div>

    <!-- Stage log -->
    <div class="mb-4 pb-4 border-b border-gray-100 dark:border-gray-700">
      <p class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3">
        {{ t('detail.stageHistory') }}
      </p>
      <StageLogList
        v-if="logs.length"
        :logs="logs"
      />
      <p
        v-else
        class="text-sm text-gray-400 dark:text-gray-500"
      >
        {{ t('detail.noStageHistory') }}
      </p>
    </div>

    <!-- Contacts -->
    <div>
      <p class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3">
        {{ t('detail.contacts') }}
      </p>
      <ul
        v-if="contacts.length || pendingContacts.length"
        class="space-y-2 mb-4 max-h-40 overflow-y-auto pr-1"
      >
        <li
          v-for="c in contacts"
          :key="c.id"
          class="flex items-start justify-between gap-2 text-sm border border-gray-100 dark:border-gray-700 rounded-lg px-3 py-2"
        >
          <div>
            <span class="font-medium text-gray-800 dark:text-gray-100">{{ c.name }}</span>
            <span
              v-if="c.role"
              class="text-xs text-gray-500 dark:text-gray-400 ml-1"
            >({{ c.role }})</span>
            <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 space-x-2">
              <span v-if="c.email">{{ c.email }}</span>
              <span v-if="c.phone">{{ c.phone }}</span>
            </div>
          </div>
          <button
            :aria-label="t('detail.removeContact')"
            :title="t('detail.removeContact')"
            class="inline-flex items-center justify-center p-1.5 rounded text-red-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-red-500 transition-colors shrink-0"
            @click="removeContact(c.id)"
          >
            <svg
              class="w-3.5 h-3.5"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.5"
              aria-hidden="true"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0"
              />
            </svg>
          </button>
        </li>
        <li
          v-for="(c, i) in pendingContacts"
          :key="'p'+i"
          class="flex items-start justify-between gap-2 text-sm border border-dashed border-blue-300 dark:border-blue-600 rounded-lg px-3 py-2 opacity-70"
        >
          <div>
            <span class="font-medium text-gray-800 dark:text-gray-100">{{ c.name }}</span>
            <span
              v-if="c.role"
              class="text-xs text-gray-500 dark:text-gray-400 ml-1"
            >({{ c.role }})</span>
            <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 space-x-2">
              <span v-if="c.email">{{ c.email }}</span>
              <span v-if="c.phone">{{ c.phone }}</span>
            </div>
          </div>
          <button
            :aria-label="t('detail.removeContact')"
            :title="t('detail.removeContact')"
            class="inline-flex items-center justify-center p-1.5 rounded text-red-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-red-500 transition-colors shrink-0"
            @click="pendingContacts.splice(i, 1)"
          >
            <svg
              class="w-3.5 h-3.5"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.5"
              aria-hidden="true"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0"
              />
            </svg>
          </button>
        </li>
      </ul>
      <p
        v-else
        class="text-sm text-gray-400 dark:text-gray-500 mb-4"
      >
        {{ t('detail.noContactsYet') }}
      </p>
      <form
        class="flex flex-col gap-2"
        @submit.prevent="addContact"
      >
        <div class="flex flex-col md:flex-row gap-2">
          <input
            v-model="newContact.name"
            :placeholder="t('common.name')"
            required
            class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
          <input
            v-model="newContact.role"
            :placeholder="t('common.role')"
            class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
        </div>
        <div class="flex flex-col md:flex-row gap-2">
          <input
            v-model="newContact.email"
            :placeholder="t('common.email')"
            type="email"
            class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
          <input
            v-model="newContact.phone"
            :placeholder="t('common.phone')"
            class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
        </div>
        <button
          type="submit"
          class="min-h-11 md:min-h-0 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
        >
          {{ t('detail.addContact') }}
        </button>
      </form>
    </div>

    <!-- Meetings -->
    <div class="mt-4 pt-4 border-t border-gray-100 dark:border-gray-700">
      <p class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3">
        {{ t('detail.meetings') }}
      </p>
      <ul
        v-if="sortedMeetings.length"
        class="space-y-2 mb-4 max-h-48 overflow-y-auto pr-1"
      >
        <li
          v-for="m in sortedMeetings"
          :key="m.id"
          :class="m.past ? 'opacity-60 border-gray-100 dark:border-gray-700' : (isUrgent(m.scheduled_at) ? 'border-amber-300 dark:border-amber-600 bg-amber-50 dark:bg-amber-900/20' : 'border-gray-100 dark:border-gray-700')"
          class="border rounded-lg px-3 py-2 text-sm"
        >
          <div
            v-if="editingMeeting && editingMeeting.id === m.id"
            class="flex flex-col gap-2"
          >
            <input
              v-model="editingMeeting.title"
              :placeholder="t('common.title')"
              required
              class="border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100"
            >
            <input
              v-model="editingMeeting.scheduled_at"
              type="datetime-local"
              required
              class="border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100"
            >
            <input
              v-model="editingMeeting.url"
              :placeholder="t('common.url')"
              type="url"
              class="border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100"
            >
            <textarea
              v-model="editingMeeting.notes"
              rows="2"
              :placeholder="t('common.notes')"
              class="border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 resize-none"
            />
            <div class="flex gap-2 justify-end">
              <button
                class="text-xs text-gray-500 dark:text-gray-400 px-2 py-1"
                @click="editingMeeting = null"
              >
                {{ t('common.cancel') }}
              </button>
              <button
                :disabled="meetingBusy"
                class="text-xs bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white px-2 py-1 rounded"
                @click="saveMeetingEdit"
              >
                {{ t('common.save') }}
              </button>
            </div>
          </div>
          <div
            v-else
            class="flex items-start justify-between gap-2"
          >
            <div>
              <div class="font-medium text-gray-800 dark:text-gray-100">
                {{ m.title }}
              </div>
              <div class="text-xs text-gray-500 dark:text-gray-400">
                {{ formatDate(m.scheduled_at) }}
              </div>
              <a
                v-if="m.url"
                :href="m.url"
                target="_blank"
                rel="noopener"
                class="text-xs text-blue-600 dark:text-blue-400 hover:underline"
              >{{ m.url }}</a>
              <p
                v-if="m.notes"
                class="text-xs text-gray-500 dark:text-gray-400 mt-0.5"
              >
                {{ m.notes }}
              </p>
            </div>
            <div class="flex gap-1 shrink-0">
              <button
                :aria-label="t('detail.editMeeting')"
                :title="t('detail.editMeeting')"
                class="inline-flex items-center justify-center p-1.5 rounded text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:text-gray-200 dark:hover:bg-gray-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors"
                @click="editMeeting(m)"
              >
                <svg
                  class="w-3.5 h-3.5"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.5"
                  aria-hidden="true"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10"
                  />
                </svg>
              </button>
              <button
                :aria-label="t('detail.deleteMeeting')"
                :title="t('detail.deleteMeeting')"
                class="inline-flex items-center justify-center p-1.5 rounded text-red-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-red-500 transition-colors"
                @click="removeMeeting(m.id)"
              >
                <svg
                  class="w-3.5 h-3.5"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.5"
                  aria-hidden="true"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0"
                  />
                </svg>
              </button>
            </div>
          </div>
        </li>
      </ul>
      <p
        v-else
        class="text-sm text-gray-400 dark:text-gray-500 mb-4"
      >
        {{ t('detail.noMeetingsYet') }}
      </p>
      <form
        class="flex flex-col gap-2"
        @submit.prevent="addMeeting"
      >
        <div class="flex flex-col md:flex-row gap-2">
          <input
            v-model="newMeeting.title"
            :placeholder="t('common.title')"
            required
            class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
          <input
            v-model="newMeeting.scheduled_at"
            type="datetime-local"
            required
            class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
        </div>
        <div class="flex flex-col md:flex-row gap-2">
          <input
            v-model="newMeeting.url"
            :placeholder="t('common.url')"
            type="url"
            class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
          <input
            v-model="newMeeting.notes"
            :placeholder="t('common.notes')"
            class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
        </div>
        <button
          type="submit"
          :disabled="meetingBusy"
          class="min-h-11 md:min-h-0 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
        >
          {{ t('detail.addMeeting') }}
        </button>
      </form>
    </div>

    <div class="flex gap-2 justify-end mt-6 pt-4 border-t border-gray-100 dark:border-gray-700">
      <button
        class="min-h-11 md:min-h-0 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 text-sm font-medium px-4 py-2 rounded-lg transition-colors"
        @click="emit('close')"
      >
        {{ t('common.cancel') }}
      </button>
      <button
        :disabled="saving"
        class="min-h-11 md:min-h-0 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
        @click="saveDetail"
      >
        {{ t('common.save') }}
      </button>
    </div>

    <!-- Manage Stages (per-job) -->
    <BaseDialog
      v-if="stagesMgmt"
      width="w-full max-w-sm"
      @close="stagesMgmt = false"
    >
      <div class="flex justify-between items-center mb-4">
        <h3 class="font-semibold text-gray-800 dark:text-gray-100">
          {{ t('detail.manageStagesTitle') }}
        </h3>
        <button
          class="min-h-11 min-w-11 inline-flex items-center justify-center md:min-h-0 md:min-w-0 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-lg leading-none"
          @click="stagesMgmt = false"
        >
          ✕
        </button>
      </div>
      <StageListEditor
        :stages="stages"
        @add="addStage"
        @rename="renameStage"
        @remove="removeStage"
        @reorder="reorderStages"
      />
    </BaseDialog>

    <StageCommentDialog
      v-if="stageComment.open"
      :is-last-stage="stageComment.isLastStage"
      :initial-status="edit.status"
      @confirm="confirmStageComment"
      @close="stageComment.open = false"
    />
  </BaseDialog>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import * as api from '../api'
import { statuses } from '../constants'
import { isoToDate, formatDate, toRFC3339, toDatetimeLocal, isUrgent } from '../utils/dates'
import { useJobs } from '../composables/useJobs'
import { useMeetings } from '../composables/useMeetings'
import { useI18n } from '../composables/useI18n'
import BaseDialog from './BaseDialog.vue'
import StageLogList from './StageLogList.vue'
import StageListEditor from './StageListEditor.vue'
import StageCommentDialog from './StageCommentDialog.vue'

const props = defineProps({
  job: { type: Object, required: true },
})

const emit = defineEmits(['close', 'saved'])

const { toggleTopMatch } = useJobs()
const { loadUpcomingMeetings } = useMeetings()
const { t, tStage } = useI18n()

const edit = ref({
  company: props.job.company, position: props.job.position, status: props.job.status,
  applied_at: isoToDate(props.job.applied_at), notes: props.job.notes, url: props.job.url,
  stage_id: props.job.stage_id,
})
const contacts = ref([])
const logs = ref([])
const stages = ref([])
const meetings = ref([])
const newContact = ref({ name: '', role: '', email: '', phone: '' })
const pendingContacts = ref([])
const deletedContactIds = ref([])
const newMeeting = ref({ title: '', scheduled_at: '', url: '', notes: '' })
const editingMeeting = ref(null)
const stagesMgmt = ref(false)
const stageComment = ref({ open: false, isLastStage: false })
const saving = ref(false) // guards saveDetail/confirmStageComment against double-submit
const meetingBusy = ref(false) // guards the add/edit meeting forms

onMounted(async () => {
  const jobId = props.job.id
  ;[contacts.value, logs.value, stages.value, meetings.value] = await Promise.all([
    api.fetchContacts(jobId),
    api.fetchLogs(jobId),
    api.fetchJobStages(jobId),
    api.fetchJobMeetings(jobId),
  ])
})

function contactRequests() {
  const jobId = props.job.id
  return [
    ...pendingContacts.value.map(c => api.addContact(jobId, c)),
    ...deletedContactIds.value.map(cid => api.deleteContact(jobId, cid)),
  ]
}

async function saveDetail() {
  const newStage = edit.value.stage_id || null
  const oldStage = props.job.stage_id || null
  if (newStage !== oldStage) {
    const isLastStage = stages.value.length > 0 && stages.value[stages.value.length - 1]?.id === newStage
    stageComment.value = { open: true, isLastStage }
    return
  }
  if (saving.value) return
  saving.value = true
  try {
    await Promise.all([
      api.updateJob(props.job.id, edit.value),
      ...contactRequests(),
    ])
  } finally {
    saving.value = false
  }
  emit('saved')
}

async function confirmStageComment({ notes, newStatus }) {
  if (saving.value) return
  saving.value = true
  const { stage_id, ...rest } = edit.value
  if (stageComment.value.isLastStage && newStatus) rest.status = newStatus
  try {
    await Promise.all([
      api.addLog(props.job.id, { stage_id: stage_id || null, notes }),
      api.updateJob(props.job.id, rest),
      ...contactRequests(),
    ])
  } finally {
    saving.value = false
  }
  stageComment.value.open = false
  emit('saved')
}

function addContact() {
  pendingContacts.value.push({ ...newContact.value })
  newContact.value = { name: '', role: '', email: '', phone: '' }
}

function removeContact(id) {
  contacts.value = contacts.value.filter(c => c.id !== id)
  deletedContactIds.value.push(id)
}

const sortedMeetings = computed(() => {
  const now = Date.now()
  const upcoming = meetings.value
    .filter(m => new Date(m.scheduled_at).getTime() >= now)
    .sort((a, b) => new Date(a.scheduled_at) - new Date(b.scheduled_at))
  const past = meetings.value
    .filter(m => new Date(m.scheduled_at).getTime() < now)
    .sort((a, b) => new Date(b.scheduled_at) - new Date(a.scheduled_at))
  return [...upcoming.map(m => ({ ...m, past: false })), ...past.map(m => ({ ...m, past: true }))]
})

async function refreshMeetings() {
  meetings.value = await api.fetchJobMeetings(props.job.id)
  await loadUpcomingMeetings()
}

async function addMeeting() {
  if (!newMeeting.value.title || !newMeeting.value.scheduled_at) return
  if (meetingBusy.value) return
  meetingBusy.value = true
  try {
    await api.addMeeting(props.job.id, {
      title: newMeeting.value.title,
      scheduled_at: toRFC3339(newMeeting.value.scheduled_at),
      url: newMeeting.value.url,
      notes: newMeeting.value.notes,
    })
  } finally {
    meetingBusy.value = false
  }
  newMeeting.value = { title: '', scheduled_at: '', url: '', notes: '' }
  await refreshMeetings()
}

function editMeeting(m) {
  editingMeeting.value = { id: m.id, title: m.title, scheduled_at: toDatetimeLocal(m.scheduled_at), url: m.url, notes: m.notes }
}

async function saveMeetingEdit() {
  const meetingEdit = editingMeeting.value
  if (!meetingEdit.title || !meetingEdit.scheduled_at) return
  if (meetingBusy.value) return
  meetingBusy.value = true
  try {
    await api.updateMeeting(props.job.id, meetingEdit.id, {
      title: meetingEdit.title,
      scheduled_at: toRFC3339(meetingEdit.scheduled_at),
      url: meetingEdit.url,
      notes: meetingEdit.notes,
    })
  } finally {
    meetingBusy.value = false
  }
  editingMeeting.value = null
  await refreshMeetings()
}

async function removeMeeting(id) {
  await api.deleteMeeting(props.job.id, id)
  await refreshMeetings()
}

async function reloadStages() {
  stages.value = await api.fetchJobStages(props.job.id)
}

async function addStage(name) {
  const maxOrder = stages.value.length > 0 ? Math.max(...stages.value.map(s => s.sort_order)) : 0
  await api.addJobStage(props.job.id, { name, sort_order: maxOrder + 1 })
  await reloadStages()
}

async function renameStage(stage, name) {
  stage.name = name
  await api.updateStage(stage)
}

async function removeStage(id) {
  await api.deleteStage(id)
  await reloadStages()
}

async function reorderStages(fromIdx, toIdx) {
  await api.swapStageOrder(stages.value[fromIdx], stages.value[toIdx])
  await reloadStages()
}
</script>
